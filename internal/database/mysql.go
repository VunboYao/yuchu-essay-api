// Package database 提供 MySQL（GORM）与 Redis 全局连接，供 handler 直接访问。
package database

import (
	"context"
	"errors"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB 全局 GORM 实例，由 main 在启动时初始化。
var DB *gorm.DB

const (
	// 连接超时时间
	mysqlPingTimeout = 2 * time.Second
	// 最大连接数
	mysqlMaxOpen = 25
	// 最大空闲连接数
	mysqlMaxIdle = 10
	// 连接最大生命周期
	mysqlConnMaxLife = time.Hour
)

// InitMySQL 建立 MySQL 连接、配置连接池、Ping 校验后赋值给 DB。
func InitMySQL(dsn string) error {
	if dsn == "" {
		return errors.New("database dsn is empty")
	}
	// 连接 MySQL
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	// 获取 SQL 连接
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	// 设置连接池
	sqlDB.SetMaxOpenConns(mysqlMaxOpen)
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(mysqlMaxIdle)
	// 设置连接最大生命周期
	sqlDB.SetConnMaxLifetime(mysqlConnMaxLife)

	// 连接超时时间
	ctx, cancel := context.WithTimeout(context.Background(), mysqlPingTimeout)
	defer cancel()
	// 连接 MySQL
	if err := sqlDB.PingContext(ctx); err != nil {
		// 关闭 SQL 连接
		_ = sqlDB.Close()
		return err
	}
	// 赋值给 DB
	DB = db
	return nil
}

// CloseMySQL 关闭全局 MySQL 连接池；进程退出前调用，重复调用安全。
func CloseMySQL() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	DB = nil
	return err
}
