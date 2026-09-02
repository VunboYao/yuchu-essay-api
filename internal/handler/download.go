// Day7：下载资料后台 CRUD 与小程序免费下载签名下发。
package handler

import (
	"errors"
	"strings"

	"course_server/internal/database"
	admindto "course_server/internal/dto/admin"
	"course_server/internal/middleware"
	"course_server/internal/model"
	"course_server/internal/pkg/response"
	"course_server/internal/service/download"

	"github.com/gin-gonic/gin"
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
	ownerIDs := collectOwnerArticleIDsFromFiles(list)
	owners := ownerArticleRefsByIDs(database.DB, ownerIDs)
	items := BuildDownloadFileListItems(list, owners)
	response.Page(c, items, total, page, pageSize)
}

// CreateDownloadFile 创建下载文件
func (s *Server) CreateDownloadFile(c *gin.Context) {
	var req admindto.CreateDownloadFileRequest
	if !bindJSON(c, &req) {
		return
	}
	file, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	s.applyDownloadFileWriteMedia(&file)
	// 校验归属文章是否存在，避免写入孤儿资料
	if err := model.CheckDownloadOwnerExists(database.DB, file.OwnerType, file.OwnerID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := CreateAndReload(database.DB, &file, file.UpdateMap()); err != nil {
		response.ServerError(c, "failed to create download file"+err.Error())
		return
	}
	middleware.SetOpLogRemark(c, "创建下载资料："+file.FileName)
	middleware.SetOpLogTargetID(c, file.ID)
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
	if !bindJSON(c, &req) {
		return
	}
	file, err := req.ValidateForUpdate(item)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	s.applyDownloadFileWriteMedia(&file)
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
	middleware.SetOpLogRemark(c, "更新下载资料："+item.FileName)
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
	middleware.SetOpLogRemark(c, "删除下载资料："+item.FileName)
	response.Success(c, gin.H{"id": id})
}

// SignDownloadURL 下发免费资料下载地址；业务规则在 download.Service。
func (s *Server) SignDownloadURL(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	out, err := s.Download.SignDownloadURL(c.Request.Context(), download.SignInput{
		FileID:      uint(id),
		UserID:      currentAuthID(c),
		UnlockToken: strings.TrimSpace(c.Query("unlockToken")),
		SourceScene: c.Query("sourceScene"),
	})
	if err != nil {
		writeDownloadErr(c, err, "failed to record download")
		return
	}
	response.Success(c, gin.H{
		"url":       out.URL,
		"expiresIn": out.ExpiresIn,
		"file":      out.File,
	})
}

// UnlockDownloadAd 激励视频看完后签发短时 unlockToken，供 sign-url 校验。
func (s *Server) UnlockDownloadAd(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	out, err := s.Download.UnlockDownloadAd(c.Request.Context(), download.UnlockInput{
		FileID:   uint(id),
		UserID:   currentAuthID(c),
		PagePath: strings.TrimSpace(c.Query("pagePath")),
	})
	if err != nil {
		writeDownloadErr(c, err, "failed to record ad unlock")
		return
	}
	response.Success(c, gin.H{
		"unlockToken": out.UnlockToken,
		"expiresIn":   out.ExpiresIn,
	})
}

// writeDownloadErr 将 service sentinel 映射为稳定对外文案；未知错误用 fallback，不拼 err.Error()。
func writeDownloadErr(c *gin.Context, err error, fallback500 string) {
	switch {
	case errors.Is(err, download.ErrNotFound):
		response.NotFound(c, download.ErrNotFound.Error())
	case errors.Is(err, download.ErrNotFree):
		response.Forbidden(c, download.ErrNotFree.Error())
	case errors.Is(err, download.ErrLoginRequired):
		response.Unauthorized(c, download.ErrLoginRequired.Error())
	case errors.Is(err, download.ErrRewardedVideoRequired):
		response.Forbidden(c, download.ErrRewardedVideoRequired.Error())
	case errors.Is(err, download.ErrInvalidUnlockToken):
		response.Forbidden(c, download.ErrInvalidUnlockToken.Error())
	case errors.Is(err, download.ErrGateNotEnabled):
		response.BadRequest(c, download.ErrGateNotEnabled.Error())
	case errors.Is(err, download.ErrGateCheck):
		response.ServerError(c, download.ErrGateCheck.Error())
	case errors.Is(err, download.ErrRecordDownload):
		response.ServerError(c, download.ErrRecordDownload.Error())
	case errors.Is(err, download.ErrRecordAdUnlock):
		response.ServerError(c, download.ErrRecordAdUnlock.Error())
	case errors.Is(err, download.ErrIssueToken):
		response.ServerError(c, download.ErrIssueToken.Error())
	default:
		response.ServerError(c, fallback500)
	}
}
