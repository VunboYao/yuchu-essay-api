// Day4：后台教材分类单元（taxonomy_units）与标签（tags）CRUD。
package handler

import (
	"course_server/internal/database"
	admindto "course_server/internal/dto/admin"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListTaxonomyUnits 分页查询教材分类单元。
func (s *Server) ListTaxonomyUnits(c *gin.Context) {
	// 分页参数
	page, pageSize := pageParams(c)
	// 数据库模型
	db := database.DB.Model(&model.TaxonomyUnit{})
	// 年级筛选
	if grade := c.Query("grade"); grade != "" {
		db = db.Where("grade = ?", grade)
	}
	// 学期筛选
	if semester := c.Query("semester"); semester != "" {
		db = db.Where("semester = ?", semester)
	}
	// 状态筛选
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	// 统计总数
	var total int64
	db.Count(&total)
	// 查询列表
	var list []model.TaxonomyUnit
	// 分页查询: grade 年级、semester 学期、unit 单元、sort_order 排序权重
	if err := paginate(db.Order("grade ASC, semester ASC, unit ASC, sort_order DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list taxonomy units")
		return
	}
	// 返回结果
	response.Page(c, list, total, page, pageSize)
}

// CreateTaxonomyUnit 创建教材分类单元
func (s *Server) CreateTaxonomyUnit(c *gin.Context) {
	// 绑定请求参数
	var req admindto.CreateTaxonomyUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验请求参数
	unit, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// sortOrder 未传时自动递增
	if req.SortOrder == nil {
		unit.SortOrder = model.NextSortOrder(database.DB, &model.TaxonomyUnit{})
	}
	// 校验分类单元重复
	if err := model.CheckTaxonomyUnitDuplicate(database.DB, unit, 0); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 写入数据库
	if err := database.DB.Create(&unit).Error; err != nil {
		response.ServerError(c, "failed to create taxonomy unit")
		return
	}
	response.Success(c, unit)
}

// UpdateTaxonomyUnit 更新教材分类单元
func (s *Server) UpdateTaxonomyUnit(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.TaxonomyUnit
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "taxonomy unit not found")
		return
	}
	// 绑定请求参数
	var req admindto.UpdateTaxonomyUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验请求参数
	unit, err := req.ValidateForUpdate(item)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验分类单元重复: 排除当前分类单元本身
	if err := model.CheckTaxonomyUnitDuplicate(database.DB, unit, item.ID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 更新数据库
	if err := UpdateAndReload(database.DB, &item, unit.UpdateMap()); err != nil {
		response.ServerError(c, "failed to update taxonomy unit")
		return
	}
	response.Success(c, item)
}

// 标签: 列表查询
func (s *Server) ListTags(c *gin.Context) {
	// 分页参数
	page, pageSize := pageParams(c)
	db := database.DB.Model(&model.Tag{})
	// 过滤 — type 精确匹配，keyword 模糊搜索
	if typ := c.Query("type"); typ != "" {
		db = db.Where("type = ?", typ)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	// 统计总数
	var total int64
	db.Count(&total)
	var list []model.Tag
	// 排序查询 — 按 sort_order DESC, id DESC 排序，即权重越大、越新创建的排在前面
	if err := paginate(db.Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list tags")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

// 标签: 创建
func (s *Server) CreateTag(c *gin.Context) {
	// 绑定请求参数
	var req admindto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验请求参数
	tag, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验标签重复
	if err := model.CheckTagNameTypeDuplicate(database.DB, tag.Name, tag.Type, 0); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 写入数据库
	if err := database.DB.Create(&tag).Error; err != nil {
		response.ServerError(c, "failed to create tag")
		return
	}
	response.Success(c, tag)
}

// 标签: 更新
func (s *Server) UpdateTag(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.Tag
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "tag not found")
		return
	}
	// 绑定请求参数
	var req admindto.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验请求参数
	tag, err := req.ValidateForUpdate(item)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验标签重复: 排除当前标签本身
	if err := model.CheckTagNameTypeDuplicate(database.DB, tag.Name, tag.Type, item.ID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 更新数据库
	if err := UpdateAndReload(database.DB, &item, tag.UpdateMap()); err != nil {
		response.ServerError(c, "failed to update tag")
		return
	}
	response.Success(c, item)
}
