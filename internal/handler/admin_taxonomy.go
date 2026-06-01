// Day4：后台教材分类单元（taxonomy_units）与标签（tags）CRUD。
package handler

import (
	"strconv"

	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (s *Server) ListTaxonomyUnits(c *gin.Context) {
	page, pageSize := pageParams(c)
	db := database.DB.Model(&model.TaxonomyUnit{})
	if grade := c.Query("grade"); grade != "" {
		db = db.Where("grade = ?", grade)
	}
	if semester := c.Query("semester"); semester != "" {
		db = db.Where("semester = ?", semester)
	}
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	var total int64
	db.Count(&total)
	var list []model.TaxonomyUnit
	if err := paginate(db.Order("grade ASC, semester ASC, unit ASC, sort_order DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list taxonomy units")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (s *Server) CreateTaxonomyUnit(c *gin.Context) {
	var req model.TaxonomyUnit
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Status == 0 {
		req.Status = model.StatusEnabled
	}
	if err := database.DB.Create(&req).Error; err != nil {
		response.ServerError(c, "failed to create taxonomy unit")
		return
	}
	response.Success(c, req)
}

func (s *Server) UpdateTaxonomyUnit(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item model.TaxonomyUnit
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "taxonomy unit not found")
		return
	}
	var req model.TaxonomyUnit
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.ID = item.ID
	if err := database.DB.Model(&item).Updates(req).Error; err != nil {
		response.ServerError(c, "failed to update taxonomy unit")
		return
	}
	response.Success(c, item)
}

func (s *Server) ListTags(c *gin.Context) {
	page, pageSize := pageParams(c)
	db := database.DB.Model(&model.Tag{})
	if typ := c.Query("type"); typ != "" {
		db = db.Where("type = ?", typ)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	db.Count(&total)
	var list []model.Tag
	if err := paginate(db.Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list tags")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (s *Server) CreateTag(c *gin.Context) {
	var req model.Tag
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Status == 0 {
		req.Status = model.StatusEnabled
	}
	if err := database.DB.Create(&req).Error; err != nil {
		response.ServerError(c, "failed to create tag")
		return
	}
	response.Success(c, req)
}

func (s *Server) UpdateTag(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item model.Tag
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "tag not found")
		return
	}
	var req model.Tag
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.ID = item.ID
	if err := database.DB.Model(&item).Updates(req).Error; err != nil {
		response.ServerError(c, "failed to update tag")
		return
	}
	response.Success(c, item)
}
