// Package middleware 提供跨域与 JWT 鉴权中间件（Day1/Day3）。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 允许管理端与小程序跨域访问；OPTIONS 预检直接返回 204。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
