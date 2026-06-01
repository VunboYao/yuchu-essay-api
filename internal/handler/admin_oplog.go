// Day3：后台操作日志查询（写入由 middleware.AdminOperationLog 统一完成）。
package handler

import (
	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListOperationLogs 分页查询操作审计，支持按管理员、目标类型筛选。
func (s *Server) ListOperationLogs(c *gin.Context) {
	page, pageSize := pageParams(c)
	db := database.DB.Model(&model.AdminOperationLog{})
	if adminID := c.Query("adminId"); adminID != "" {
		db = db.Where("admin_id = ?", adminID)
	}
	if targetType := c.Query("targetType"); targetType != "" {
		db = db.Where("target_type = ?", targetType)
	}
	var total int64
	db.Count(&total)
	var list []model.AdminOperationLog
	if err := paginate(db.Order("id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list operation logs")
		return
	}
	response.Page(c, list, total, page, pageSize)
}
