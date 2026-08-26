// Day4：后台教材分类单元（taxonomy_units）与标签（tags）CRUD。
package handler

import (
	"fmt"

	"course_server/internal/database"
	admindto "course_server/internal/dto/admin"
	"course_server/internal/middleware"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// taxonomyUnitLabel 生成教材分类单元标签
func taxonomyUnitLabel(unit model.TaxonomyUnit) string {
	return fmt.Sprintf("%d年级%s册第%d单元", unit.Grade, unit.Semester, unit.Unit)
}

// ListTaxonomyUnits 分页查询教材分类单元。
func (s *Server) ListTaxonomyUnits(c *gin.Context) {
	// 分页参数
	page, pageSize := pageParams(c)
	// 数据库模型: 教材分类单元
	db := database.DB.Model(&model.TaxonomyUnit{})
	// 年级筛选: 年级
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
	if !bindJSON(c, &req) {
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
	middleware.SetOpLogRemark(c, "创建分类单元："+taxonomyUnitLabel(unit))
	middleware.SetOpLogTargetID(c, unit.ID)
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
	if !bindJSON(c, &req) {
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
	middleware.SetOpLogRemark(c, "更新分类单元："+taxonomyUnitLabel(item))
	response.Success(c, item)
}

// DeleteTaxonomyUnit 物理删除教材分类单元；须已禁用且无文章占用同一四元组。
func (s *Server) DeleteTaxonomyUnit(c *gin.Context) {
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
	if err := model.CheckTaxonomyUnitDeletable(database.DB, item); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := database.DB.Delete(&item).Error; err != nil {
		response.ServerError(c, "failed to delete taxonomy unit")
		return
	}
	middleware.SetOpLogRemark(c, "删除分类单元："+taxonomyUnitLabel(item))
	response.Success(c, gin.H{"id": id})
}

// 标签: 列表查询
func (s *Server) ListTags(c *gin.Context) {
	// 分页参数
	page, pageSize := pageParams(c)
	db := database.DB.Model(&model.Tag{})
	// keyword 模糊搜索标签名
	if keyword := c.Query("keyword"); keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	// status 筛选：选用标签时传 1 仅返回启用项
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
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
	if !bindJSON(c, &req) {
		return
	}
	// 校验请求参数
	tag, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验标签名重复
	if err := model.CheckTagNameDuplicate(database.DB, tag.Name, 0); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 写入数据库
	if err := database.DB.Create(&tag).Error; err != nil {
		response.ServerError(c, "failed to create tag")
		return
	}
	middleware.SetOpLogRemark(c, "创建标签："+tag.Name)
	middleware.SetOpLogTargetID(c, tag.ID)
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
	if !bindJSON(c, &req) {
		return
	}
	// 校验请求参数
	tag, err := req.ValidateForUpdate(item)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 校验标签名重复: 排除当前标签本身
	if err := model.CheckTagNameDuplicate(database.DB, tag.Name, item.ID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 启用→禁用：须无文章/作文关联，避免内容页标签名丢失
	if item.Status == model.StatusEnabled && tag.Status == model.StatusDisabled {
		if err := model.CheckTagDisableAllowed(database.DB, item.ID); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}
	// 更新数据库
	if err := UpdateAndReload(database.DB, &item, tag.UpdateMap()); err != nil {
		response.ServerError(c, "failed to update tag")
		return
	}
	middleware.SetOpLogRemark(c, "更新标签："+item.Name)
	response.Success(c, item)
}

// DeleteTag 物理删除标签；须已禁用且无文章/作文关联。
func (s *Server) DeleteTag(c *gin.Context) {
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
	if err := model.CheckTagDeletable(database.DB, item); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := database.DB.Delete(&item).Error; err != nil {
		response.ServerError(c, "failed to delete tag")
		return
	}
	middleware.SetOpLogRemark(c, "删除标签："+item.Name)
	response.Success(c, gin.H{"id": id})
}
