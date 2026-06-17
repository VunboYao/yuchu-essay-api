package testutil

import (
	"testing"

	"course_server/internal/config"
	"course_server/internal/router"

	"github.com/gin-gonic/gin"
)

// TestConfig 返回固定 JWT 配置的测试用 Config。
func TestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			Issuer: "essay-mvp-test",
		},
	}
}

// SetupRouter 返回 gin 测试引擎（复用生产路由注册）。
func SetupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return router.Setup(TestConfig())
}
