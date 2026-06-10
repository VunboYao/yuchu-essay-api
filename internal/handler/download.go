// Day7：下载资料后台 CRUD 与小程序免费下载签名下发。
package handler

import (
	"strconv"
	"strings"
	"time"

	"course_server/internal/database"
	admindto "course_server/internal/dto/admin"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (s *Server) ListDownloadFiles(c *gin.Context) {
	// 1. 获取分页参数
	page, pageSize := pageParams(c)
	// 2. 构建查询条件
	db := database.DB.Model(&model.DownloadFile{})
	if ownerType := c.Query("ownerType"); ownerType != "" {
		db = db.Where("owner_type = ?", ownerType)
	}
	if ownerID := c.Query("ownerId"); ownerID != "" {
		db = db.Where("owner_id = ?", ownerID)
	}
	if accessType := c.Query("accessType"); accessType != "" {
		db = db.Where("access_type = ?", accessType)
	}
	var total int64
	db.Count(&total)
	var list []model.DownloadFile
	// 3. 执行查询
	if err := paginate(db.Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list download files")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

// CreateDownloadFile 创建下载文件
func (s *Server) CreateDownloadFile(c *gin.Context) {
	var req admindto.CreateDownloadFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	file, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验归属文章是否存在，避免写入孤儿资料
	if err := model.CheckDownloadOwnerExists(database.DB, file.OwnerType, file.OwnerID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := database.DB.Create(&file).Error; err != nil {
		response.ServerError(c, "failed to create download file"+err.Error())
		return
	}
	response.Success(c, file)
}

// UpdateDownloadFile 更新下载文件
func (s *Server) UpdateDownloadFile(c *gin.Context) {
	// 1. 获取 ID
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.DownloadFile
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "download file not found")
		return
	}
	var req admindto.UpdateDownloadFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	file, err := req.ValidateForUpdate(item)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验归属文章是否存在，避免更新孤儿资料
	if err := model.CheckDownloadOwnerExists(database.DB, file.OwnerType, file.OwnerID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 更新数据库
	if err := UpdateAndReload(database.DB, &item, file.UpdateMap()); err != nil {
		response.ServerError(c, "failed to update download file")
		return
	}
	response.Success(c, item)
}

func (s *Server) DeleteDownloadFile(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.DownloadFile
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "download file not found")
		return
	}
	// 已有下载记录时拒绝物理删除，避免丢统计；运营下架请用 PUT 改 status 或去掉「免费」标签
	force := c.Query("confirm") == "true"
	if err := model.CheckDownloadFileDeletable(database.DB, item, force); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := database.DB.Delete(&item).Error; err != nil {
		response.ServerError(c, "failed to delete download file")
		return
	}
	response.Success(c, gin.H{"id": id})
}

// SignDownloadURL 下发免费资料下载地址（公开接口，登录可选记录 user_id）。
// 须同时满足：status=启用、access_type=free、tags 含「免费」；否则 403。
// 成功时事务写入 download_logs 并递增 download_count（owner_type=article 时同步文章计数）。
func (s *Server) SignDownloadURL(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var file model.DownloadFile
	if err := database.DB.First(&file, id).Error; err != nil {
		response.NotFound(c, "download file not found")
		return
	}
	if !isPublicFreeDownload(file) {
		response.Forbidden(c, "download file is not free")
		return
	}

	userID := currentAuthID(c)
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.DownloadLog{
			UserID:      userID,
			FileID:      file.ID,
			OwnerType:   file.OwnerType,
			OwnerID:     file.OwnerID,
			AccessType:  file.AccessType,
			SourceScene: c.Query("sourceScene"),
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&file).UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error; err != nil {
			return err
		}
		if file.OwnerType == "article" {
			if err := tx.Model(&model.Article{}).Where("id = ?", file.OwnerID).UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, "failed to record download")
		return
	}

	response.Success(c, gin.H{
		"url":       pseudoSignedURL(file.FileURL),
		"expiresIn": 300,
		"file":      file,
	})
}

// isPublicFreeDownload 一期 MVP 免费下载底线，purchase_required/member_only/internal 均不满足。
func isPublicFreeDownload(file model.DownloadFile) bool {
	return file.Status == model.StatusEnabled &&
		file.AccessType == model.AccessTypeFree &&
		strings.Contains(file.Tags, "免费")
}

// pseudoSignedURL 开发期伪签名（拼接 expires）；生产应替换为 OSS 预签名 URL。
func pseudoSignedURL(fileURL string) string {
	sep := "?"
	if strings.Contains(fileURL, "?") {
		sep = "&"
	}
	return fileURL + sep + "expires=" + strconv.FormatInt(time.Now().Add(5*time.Minute).Unix(), 10)
}
