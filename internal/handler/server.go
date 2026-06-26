// Package handler 实现全部 HTTP 业务逻辑（Handler 直连 GORM，无 service 层）。
// 方法挂在 Server 上，由 router 注册；配置通过 Server.Config 读取 JWT/OSS/微信等。
package handler

import (
	"course_server/internal/config"
	"course_server/internal/pkg/ossutil"
)

// Server 为 Gin handler 容器，持有运行时配置。
type Server struct {
	Config *config.Config
	OSS    *ossutil.Client
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		Config: cfg,
		OSS:    ossutil.NewClient(cfg.OSS),
	}
}
