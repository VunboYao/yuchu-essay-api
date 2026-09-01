// 应用入口（Day1）：加载配置 → 初始化 MySQL/Redis → 自动迁移与种子数据 → 启动 Gin。
package main

import (
	"log"

	"course_server/internal/config"
	"course_server/internal/database"
	"course_server/internal/router"
)

func main() {
	// 1. 读取 config.yaml；文件不存在时使用 config.Load 内置的默认值
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 2. 初始化 MySQL
	if err := database.InitMySQL(cfg.Database.DSN); err != nil {
		log.Fatalf("init mysql: %v", err)
	}
	// 3. 关闭 MySQL 连接
	defer func() {
		if err := database.CloseMySQL(); err != nil {
			log.Printf("close mysql: %v", err)
		}
	}()
	// 4. Redis 连接失败仅告警，不阻断启动（一期业务尚未使用 RDB）
	if err := database.InitRedis(cfg.Redis); err != nil {
		log.Printf("redis disabled: %v", err)
	}
	// 5. 自动建表
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}
	// 6. 种子数据：默认管理员、一期固定广告位、搜索热门词
	if err := database.SeedAdmin(); err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	if err := database.SeedAdSlots(); err != nil {
		log.Fatalf("seed ad slots: %v", err)
	}
	if err := database.SeedHotSearchKeywords(); err != nil {
		log.Fatalf("seed hot search keywords: %v", err)
	}

	// 7. 启动 HTTP 服务
	r := router.Setup(cfg)
	if err := r.Run(cfg.Server.Addr); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
