// Day6：后台优秀作文（essay_examples）CRUD；通过 article_id 关联专题，标签关联见 admin_tag_relation.go。
package handler

import (
	"strconv"

	"course_server/internal/database"
	admindto "course_server/internal/dto/admin"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (s *Server) ListEssayExamples(c *gin.Context) {
	page, pageSize := pageParams(c)
	db := database.DB.Model(&model.EssayExample{})
	if grade := c.Query("grade"); grade != "" {
		db = db.Where("grade = ?", grade)
	}
	if articleID := c.Query("articleId"); articleID != "" {
		db = db.Where("article_id = ?", articleID)
	}
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		db = db.Where("title LIKE ? OR body LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	db.Count(&total)
	var list []model.EssayExample
	if err := paginate(db.Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list essay examples")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (s *Server) CreateEssayExample(c *gin.Context) {
	var req admindto.CreateEssayExampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	essay, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := model.CheckEssayArticleExists(database.DB, essay.ArticleID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := database.DB.Create(&essay).Error; err != nil {
		response.ServerError(c, "failed to create essay example")
		return
	}
	response.Success(c, essay)
}

func (s *Server) UpdateEssayExample(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item model.EssayExample
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "essay example not found")
		return
	}
	var req admindto.UpdateEssayExampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	essay, err := req.ValidateForUpdate(item)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := model.CheckEssayArticleExists(database.DB, essay.ArticleID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := UpdateAndReload(database.DB, &item, essay.UpdateMap()); err != nil {
		response.ServerError(c, "failed to update essay example")
		return
	}
	response.Success(c, item)
}

func (s *Server) DeleteEssayExample(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := database.DB.Delete(&model.EssayExample{}, id).Error; err != nil {
		response.ServerError(c, "failed to delete essay example")
		return
	}
	response.Success(c, gin.H{"id": id})
}
