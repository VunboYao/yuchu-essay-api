package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"course_server/internal/database"
	"course_server/internal/model"

	"github.com/gin-gonic/gin"
)

const (
	ctxOpLogRemark   = "oplog_remark"
	ctxOpLogTargetID = "oplog_target_id"
)

// SetOpLogRemark 由 handler 在写操作成功后调用，供审计中间件写入 remark。
func SetOpLogRemark(c *gin.Context, remark string) {
	c.Set(ctxOpLogRemark, remark)
}

// SetOpLogTargetID 供 POST 创建类接口补充 targetId（路径无 :id 时中间件默认为 0）。
func SetOpLogTargetID(c *gin.Context, id uint) {
	c.Set(ctxOpLogTargetID, id)
}

// AdminOperationLog 统一记录后台写操作审计（Day3）。
// 仅记录 POST/PUT/DELETE 且响应 2xx 的请求；GET、预检与失败请求不写入。
// 审计为旁路操作，写入失败不影响主流程（忽略错误）。
func AdminOperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodOptions {
			return
		}
		if c.Writer.Status() < http.StatusOK || c.Writer.Status() >= http.StatusMultipleChoices {
			return
		}
		adminID, _ := c.Get("auth_id")
		id, _ := adminID.(uint)
		if id == 0 {
			return
		}
		targetID := 0
		if raw := c.Param("id"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				targetID = parsed
			}
		}
		if targetID == 0 {
			if ctxID, ok := c.Get(ctxOpLogTargetID); ok {
				if v, ok := ctxID.(uint); ok && v > 0 {
					targetID = int(v)
				}
			}
		}
		log := model.AdminOperationLog{
			AdminID:    id,
			Action:     c.Request.Method + " " + c.FullPath(),
			TargetType: model.AdminOpLogTargetTypeFromPath(c.FullPath()),
			TargetID:   uint(targetID),
		}
		if remark, ok := c.Get(ctxOpLogRemark); ok {
			if s, ok := remark.(string); ok {
				log.Remark = strings.TrimSpace(s)
			}
		}
		database.DB.Create(&log)
	}
}
