package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"course_server/internal/database"
	"course_server/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminOperationLog 统一记录后台写操作审计（Day3）。
// 仅记录 POST/PUT/DELETE 且响应 2xx 的请求；GET、预检与失败请求不写入。
// 审计为旁路操作，写入失败不影响主流程（忽略错误）。
func AdminOperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 如果请求方法为 GET 或 OPTIONS，则不记录审计日志
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodOptions {
			return
		}
		// 如果响应状态码小于 200 或大于等于 300，则不记录审计日志
		if c.Writer.Status() < http.StatusOK || c.Writer.Status() >= http.StatusMultipleChoices {
			return
		}
		// 获取管理员ID
		adminID, _ := c.Get("auth_id")
		id, _ := adminID.(uint)
		// 如果管理员ID为0，则不记录审计日志
		if id == 0 {
			return
		}
		// 获取目标ID
		targetID := 0
		// 如果路径参数 id 存在，则获取目标ID
		if raw := c.Param("id"); raw != "" {
			if id, err := strconv.Atoi(raw); err == nil && id > 0 {
				targetID = id
			}
		}
		// 写入审计日志
		database.DB.Create(&model.AdminOperationLog{
			AdminID:    id,
			Action:     c.Request.Method + " " + c.FullPath(),
			TargetType: resourceFromPath(c.FullPath()),
			TargetID:   uint(targetID),
		})
	}
}

// resourceFromPath 取 /admin/api/v1/<resource>... 中的 resource 段作为审计目标类型。
func resourceFromPath(fullPath string) string {
	const prefix = "/admin/api/v1/"
	if !strings.HasPrefix(fullPath, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(fullPath, prefix)
	if idx := strings.IndexByte(rest, '/'); idx >= 0 {
		return rest[:idx]
	}
	return rest
}
