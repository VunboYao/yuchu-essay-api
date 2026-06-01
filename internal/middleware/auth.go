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
	return auth(cfg, "admin")
}

// UserAuth 校验小程序 JWT（type=user），admin 与 user token 不可混用。
func UserAuth(cfg *config.Config) gin.HandlerFunc {
	return auth(cfg, "user")
}

func auth(cfg *config.Config, tokenType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := bearerToken(c.GetHeader("Authorization"))
		if tokenString == "" {
			response.Unauthorized(c, "missing authorization token")
			c.Abort()
			return
		}
		claims, err := jwtutil.Parse(cfg.JWT.Secret, tokenString, tokenType)
		if err != nil {
			response.Unauthorized(c, "invalid authorization token")
			c.Abort()
			return
		}
		c.Set("auth_id", claims.ID)
		c.Set("auth_subject", claims.Subject)
		c.Set("auth_role", claims.Role)
		c.Next()
	}
}

func bearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
