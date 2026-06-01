// Day9：小程序首页、分类树、已发布文章/优秀作文、搜索（仅 status=published/enabled）。
package handler

import (
	"strconv"

	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Home 聚合推荐文章、热门文章、优秀作文与启用中的首页区块配置。
func (s *Server) Home(c *gin.Context) {
	var recommended []model.Article
	database.DB.Where("status = ? AND is_recommended = ?", model.ArticleStatusPublished, true).
		Omit("body_images").Order("sort_order DESC, id DESC").Limit(10).Find(&recommended)

	var hot []model.Article
	database.DB.Where("status = ? AND is_hot = ?", model.ArticleStatusPublished, true).
		Omit("body_images").Order("view_count DESC, id DESC").Limit(10).Find(&hot)

	var essays []model.EssayExample
	database.DB.Where("status = ?", model.StatusEnabled).
		Order("sort_order DESC, id DESC").Limit(10).Find(&essays)

	var blocks []model.HomeBlock
	database.DB.Where("status = ?", model.StatusEnabled).
		Order("sort_order DESC, id DESC").Find(&blocks)

	response.Success(c, gin.H{
		"recommendedArticles": recommended,
		"hotArticles":         hot,
		"essayExamples":       essays,
		"homeBlocks":          blocks,
	})
}

func (s *Server) Categories(c *gin.Context) {
	var list []model.TaxonomyUnit
	if err := database.DB.Where("status = ?", model.StatusEnabled).
		Order("grade ASC, semester ASC, unit ASC, sort_order DESC").Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list categories")
		return
	}
	response.Success(c, list)
}

func (s *Server) PublicArticles(c *gin.Context) {
	page, pageSize := pageParams(c)
	db := articleFilters(database.DB.Model(&model.Article{}), c).Where("status = ?", model.ArticleStatusPublished)
	var total int64
	db.Count(&total)
	var list []model.Article
	if err := paginate(db.Omit("body_images").Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list articles")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

// PublicArticleDetail 已发布文章详情；递增 view_count，并返回该文章下可免费下载的资料列表。
func (s *Server) PublicArticleDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item model.Article
	if err := database.DB.Preload("Essays", "status = ?", model.StatusEnabled).
		Where("status = ?", model.ArticleStatusPublished).First(&item, id).Error; err != nil {
		response.NotFound(c, "article not found")
		return
	}
	database.DB.Model(&item).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	var downloads []model.DownloadFile
	database.DB.Where("owner_type = ? AND owner_id = ? AND status = ? AND access_type = ? AND tags LIKE ?",
		"article", item.ID, model.StatusEnabled, model.AccessTypeFree, "%免费%").
		Order("sort_order DESC, id DESC").Find(&downloads)

	response.Success(c, gin.H{
		"article":       item,
		"freeDownloads": downloads,
	})
}

func (s *Server) PublicEssays(c *gin.Context) {
	page, pageSize := pageParams(c)
	db := database.DB.Model(&model.EssayExample{}).Where("status = ?", model.StatusEnabled)
	if grade := c.Query("grade"); grade != "" {
		db = db.Where("grade = ?", grade)
	}
	if articleID := c.Query("articleId"); articleID != "" {
		db = db.Where("article_id = ?", articleID)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		db = db.Where("title LIKE ? OR body LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	db.Count(&total)
	var list []model.EssayExample
	if err := paginate(db.Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list essays")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (s *Server) PublicEssayDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item model.EssayExample
	if err := database.DB.Where("status = ?", model.StatusEnabled).First(&item, id).Error; err != nil {
		response.NotFound(c, "essay not found")
		return
	}
	database.DB.Model(&item).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	response.Success(c, item)
}

// Search 在已发布文章中检索，并写入 search_logs（供后台热门搜索统计）。
func (s *Server) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.BadRequest(c, "keyword is required")
		return
	}
	page, pageSize := pageParams(c)
	db := database.DB.Model(&model.Article{}).Where("status = ?", model.ArticleStatusPublished).
		Where("title LIKE ? OR essay_topic LIKE ? OR summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	var total int64
	db.Count(&total)
	var list []model.Article
	if err := paginate(db.Omit("body_images").Order("view_count DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to search")
		return
	}
	database.DB.Create(&model.SearchLog{Keyword: keyword, ResultCount: int(total), SourceScene: c.Query("sourceScene")})
	response.Page(c, list, total, page, pageSize)
}
