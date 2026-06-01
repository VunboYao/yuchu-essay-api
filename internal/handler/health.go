// Day1：服务健康检查。
package handler

import (
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// Health 供负载均衡与部署脚本探测，不访问数据库。
func (s *Server) Health(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok"})
}
