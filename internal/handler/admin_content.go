// Day5：后台作文专题（articles）CRUD；正文为 bodyImages 有序图片列表。
// 标签关联见 admin_tag_relation.go；上架自动维护 published_at（见 UpdateArticleStatus）。
package handler

import (
	"time"

	"course_server/internal/database"
	admindto "course_server/internal/dto/admin"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListArticles 列表
func (s *Server) ListArticles(c *gin.Context) {
	// 分页参数
	page, pageSize := pageParams(c)
	// 筛选参数
	db := articleFilters(database.DB.Model(&model.Article{}), c)
	var total int64
	// 总条数
	db.Count(&total)
	var list []model.Article
	// 排除 body_images, 按照 sort_order DESC, id DESC 排序
	if err := paginate(db.Omit("body_images").Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list articles")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

// CreateArticle 创建
func (s *Server) CreateArticle(c *gin.Context) {
	// 绑定请求体
	var req admindto.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验入参
	article, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 检查 slug 是否重复
	if err := model.CheckArticleSlugDuplicate(database.DB, article.Slug, 0); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 事务内创建文章并可选写入标签关联
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&article).Error; err != nil {
			return err
		}
		if req.TagIDs != nil {
			return replaceArticleTags(tx, article.ID, req.TagIDs)
		}
		return nil
	}); err != nil {
		response.ServerError(c, "failed to create article")
		return
	}
	response.Success(c, gin.H{
		"article": article,
		"tags":    s.articleTags(article.ID),
	})
}

// GetArticle 获取文章
func (s *Server) GetArticle(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.Article
	if err := database.DB.Preload("Essays", "status = ?", model.StatusEnabled).First(&item, id).Error; err != nil {
		response.NotFound(c, "article not found")
		return
	}
	response.Success(c, item)
}

// UpdateArticle 更新
func (s *Server) UpdateArticle(c *gin.Context) {
	// 解析请求参数id并检查
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.Article
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "article not found")
		return
	}
	// 绑定请求参数
	var req admindto.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验入参
	article, err := req.ValidateForUpdate(item)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 检查 slug 是否重复
	if err := model.CheckArticleSlugDuplicate(database.DB, article.Slug, item.ID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 事务内更新文章并可选全量替换标签
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := UpdateAndReload(tx, &item, article.UpdateMap()); err != nil {
			return err
		}
		if req.TagIDs != nil {
			return replaceArticleTags(tx, item.ID, req.TagIDs)
		}
		return nil
	}); err != nil {
		response.ServerError(c, "failed to update article:"+err.Error())
		return
	}
	response.Success(c, gin.H{
		"article": item,
		"tags":    s.articleTags(item.ID),
	})
}

// UpdateArticleStatus 上架/下架。
// 首次上架（published_at 为空）记录当前时间为发布时间，重复上架保留原值；下架则清空 published_at。
func (s *Server) UpdateArticleStatus(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	var item model.Article
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "article not found")
		return
	}

	updates := map[string]interface{}{"status": req.Status}
	if req.Status == model.ArticleStatusPublished {
		if item.PublishedAt == nil {
			updates["published_at"] = time.Now()
		}
	} else {
		updates["published_at"] = nil
	}
	if err := database.DB.Model(&item).Updates(updates).Error; err != nil {
		response.ServerError(c, "failed to update article status")
		return
	}

	database.DB.First(&item, id)
	response.Success(c, gin.H{"id": id, "status": item.Status, "publishedAt": item.PublishedAt})
}

// articleFilters 后台/公开列表共用的年级、册次、单元、类型、状态、关键词筛选。
func articleFilters(db *gorm.DB, c *gin.Context) *gorm.DB {
	// 年级筛选
	if grade := c.Query("grade"); grade != "" {
		db = db.Where("grade = ?", grade)
	}
	// 学期筛选
	if semester := c.Query("semester"); semester != "" {
		db = db.Where("semester = ?", semester)
	}
	// 单元筛选
	if unit := c.Query("unit"); unit != "" {
		db = db.Where("unit = ?", unit)
	}
	// 类型筛选
	if contentType := c.Query("contentType"); contentType != "" {
		db = db.Where("content_type = ?", contentType)
	}
	// 状态筛选
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	// 关键词筛选
	if keyword := c.Query("keyword"); keyword != "" {
		db = db.Where("title LIKE ? OR essay_topic LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	return db
}
