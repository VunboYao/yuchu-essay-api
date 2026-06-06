// Day5：后台作文专题（articles）CRUD；正文为 bodyImages 有序图片列表。
// 标签关联见 admin_tag_relation.go；上架自动维护 published_at（见 UpdateArticleStatus）。
package handler

import (
	"strconv"
	"time"

	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (s *Server) ListArticles(c *gin.Context) {
	page, pageSize := pageParams(c)
	db := articleFilters(database.DB.Model(&model.Article{}), c)
	var total int64
	db.Count(&total)
	var list []model.Article
	if err := paginate(db.Omit("body_images").Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list articles")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (s *Server) CreateArticle(c *gin.Context) {
	var req model.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	article, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := model.CheckArticleSlugDuplicate(database.DB, article.Slug, 0); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := database.DB.Create(&article).Error; err != nil {
		response.ServerError(c, "failed to create article")
		return
	}
	response.Success(c, article)
}

func (s *Server) GetArticle(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item model.Article
	if err := database.DB.Preload("Essays", "status = ?", model.StatusEnabled).First(&item, id).Error; err != nil {
		response.NotFound(c, "article not found")
		return
	}
	response.Success(c, item)
}

func (s *Server) UpdateArticle(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item model.Article
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "article not found")
		return
	}
	var req model.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	article, err := req.ValidateForUpdate(item)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := model.CheckArticleSlugDuplicate(database.DB, article.Slug, item.ID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := UpdateAndReload(database.DB, &item, article.UpdateMap()); err != nil {
		response.ServerError(c, "failed to update article")
		return
	}
	response.Success(c, item)
}

// UpdateArticleStatus 上架/下架。
// 首次上架（published_at 为空）记录当前时间为发布时间，重复上架保留原值；下架则清空 published_at。
func (s *Server) UpdateArticleStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
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
	if grade := c.Query("grade"); grade != "" {
		db = db.Where("grade = ?", grade)
	}
	if semester := c.Query("semester"); semester != "" {
		db = db.Where("semester = ?", semester)
	}
	if unit := c.Query("unit"); unit != "" {
		db = db.Where("unit = ?", unit)
	}
	if contentType := c.Query("contentType"); contentType != "" {
		db = db.Where("content_type = ?", contentType)
	}
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		db = db.Where("title LIKE ? OR essay_topic LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	return db
}
