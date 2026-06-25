// Day6：后台优秀作文（essay_examples）CRUD；article_id 可选关联专题，标签关联见 admin_tag_relation.go。
package handler

import (
	"course_server/internal/database"
	admindto "course_server/internal/dto/admin"
	"course_server/internal/middleware"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListEssayExamples 列表查询（含每行 tags；支持 tagId 单标签筛选）。
func (s *Server) ListEssayExamples(c *gin.Context) {
	// 分页参数
	page, pageSize := pageParams(c)
	db := essayFilters(database.DB.Model(&model.EssayExample{}), c)
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	var total int64
	db.Count(&total)
	var list []model.EssayExample
	if err := paginate(db.Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list essay examples")
		return
	}
	essayIDs := make([]uint, 0, len(list))
	for _, item := range list {
		essayIDs = append(essayIDs, item.ID)
	}
	items := BuildEssayListItems(list, essayTagsByEssayIDs(essayIDs))
	response.Page(c, items, total, page, pageSize)
}

// GetEssayExample 获取优秀作文详情（含关联标签）。
func (s *Server) GetEssayExample(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.EssayExample
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "essay example not found")
		return
	}
	response.Success(c, gin.H{
		"essay": item,
		"tags":  s.essayTags(item.ID),
	})
}

// CreateEssayExample 创建优秀作文。
func (s *Server) CreateEssayExample(c *gin.Context) {
	var req admindto.CreateEssayExampleRequest
	if !bindJSON(c, &req) {
		return
	}
	// 校验入参
	essay, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 检查专题是否存在
	if err := model.CheckEssayArticleExists(database.DB, essay.ArticleID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 事务内创建优秀作文并可选写入标签关联
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&essay).Error; err != nil {
			return err
		}
		if req.TagIDs != nil {
			return replaceEssayTags(tx, essay.ID, req.TagIDs)
		}
		return nil
	}); err != nil {
		response.ServerError(c, "failed to create essay example")
		return
	}
	middleware.SetOpLogRemark(c, "创建优秀作文：《"+essay.Title+"》")
	middleware.SetOpLogTargetID(c, essay.ID)
	response.Success(c, gin.H{
		"essay": essay,
		"tags":  s.essayTags(essay.ID),
	})
}

func (s *Server) UpdateEssayExample(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.EssayExample
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "essay example not found")
		return
	}
	var req admindto.UpdateEssayExampleRequest
	if !bindJSON(c, &req) {
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
	// 事务内更新优秀作文并可选全量替换标签
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := UpdateAndReload(tx, &item, essay.UpdateMap()); err != nil {
			return err
		}
		if req.TagIDs != nil {
			return replaceEssayTags(tx, item.ID, req.TagIDs)
		}
		return nil
	}); err != nil {
		response.ServerError(c, "failed to update essay example")
		return
	}
	middleware.SetOpLogRemark(c, "更新优秀作文：《"+item.Title+"》")
	response.Success(c, gin.H{
		"essay": item,
		"tags":  s.essayTags(item.ID),
	})
}

func (s *Server) DeleteEssayExample(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.EssayExample
	if err := database.DB.First(&item, id).Error; err != nil {
		response.NotFound(c, "essay example not found")
		return
	}
	if err := database.DB.Delete(&item).Error; err != nil {
		response.ServerError(c, "failed to delete essay example")
		return
	}
	middleware.SetOpLogRemark(c, "删除优秀作文：《"+item.Title+"》")
	response.Success(c, gin.H{"id": id})
}
