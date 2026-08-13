// Package handler 实现 HTTP 入口：解析入参、调 service、写响应。
// 下载签名/解锁已抽到 internal/service/download；其余模块仍为 Handler 直连 GORM。
package handler

import (
	"course_server/internal/config"
	"course_server/internal/database"
	"course_server/internal/pkg/ossutil"
	"course_server/internal/service/download"
)

// Server 为 Gin handler 容器，持有运行时配置与已抽离的 service。
type Server struct {
	Config   *config.Config
	OSS      *ossutil.Client
	Download *download.Service
}

func NewServer(cfg *config.Config) *Server {
	oss := ossutil.NewClient(cfg.OSS)
	return &Server{
		Config:   cfg,
		OSS:      oss,
		Download: download.New(database.DB, cfg.JWT.Secret, cfg.JWT.Issuer, oss),
	}
}
