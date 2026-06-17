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

// OptionalUserAuth 可选小程序 JWT：有合法 Bearer 时注入 auth_id，无 token 或无效 token 仍放行。
func OptionalUserAuth(cfg *config.Config) gin.HandlerFunc {
	return optionalAuth(cfg, "user")
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
		setAuthContext(c, claims)
		c.Next()
	}
}

func optionalAuth(cfg *config.Config, tokenType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := bearerToken(c.GetHeader("Authorization"))
		if tokenString != "" {
			if claims, err := jwtutil.Parse(cfg.JWT.Secret, tokenString, tokenType); err == nil {
				setAuthContext(c, claims)
			}
		}
		c.Next()
	}
}

func setAuthContext(c *gin.Context, claims *jwtutil.Claims) {
	c.Set("auth_id", claims.ID)
	c.Set("auth_subject", claims.Subject)
	c.Set("auth_role", claims.Role)
}

// bearerToken 从 Authorization 提取 JWT，兼容 Bearer <token> 与裸 token 两种写法。
func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return header
}
