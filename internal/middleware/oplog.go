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
		targetID, _ := strconv.Atoi(c.Param("id"))
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
