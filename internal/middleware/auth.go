package middleware

import (
	"strings"

	"course_server/internal/config"
	jwtutil "course_server/internal/pkg/jwt"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminAuth 校验后台 JWT（type=admin），通过后注入 auth_id / auth_subject / auth_role。
func AdminAuth(cfg *config.Config) gin.HandlerFunc {
	// 校验管理员 JWT
	return auth(cfg, "admin")
}

// UserAuth 校验小程序 JWT（type=user），admin 与 user token 不可混用。
func UserAuth(cfg *config.Config) gin.HandlerFunc {
	// 校验用户 JWT
	return auth(cfg, "user")
}

func auth(cfg *config.Config, tokenType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取 Authorization Header 中的 Bearer Token
		tokenString := bearerToken(c.GetHeader("Authorization"))
		if tokenString == "" {
			response.Unauthorized(c, "missing authorization token")
			c.Abort()
			return
		}
		// 解析 JWT
		claims, err := jwtutil.Parse(cfg.JWT.Secret, tokenString, tokenType)
		if err != nil {
			response.Unauthorized(c, "invalid authorization token")
			c.Abort()
			return
		}
		// 存储用户ID
		c.Set("auth_id", claims.ID)
		// 存储用户名
		c.Set("auth_subject", claims.Subject)
		// 存储用户角色
		c.Set("auth_role", claims.Role)
		// 继续处理请求
		c.Next()
	}
}

// 提取 Bearer Token
func bearerToken(header string) string {
	// 分割 header 为两个部分
	parts := strings.SplitN(header, " ", 2)
	// 如果长度为 2 且第一个部分为 Bearer，则返回第二个部分
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
