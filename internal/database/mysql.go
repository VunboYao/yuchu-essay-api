// Package database 提供 MySQL（GORM）与 Redis 全局连接，供 handler 直接访问。
package database

import (
	"errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB 全局 GORM 实例，由 main 在启动时初始化。
var DB *gorm.DB

// InitMySQL 建立 MySQL 连接并赋值给 DB。
func InitMySQL(dsn string) error {
	if dsn == "" {
		return errors.New("database dsn is empty")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db
	return nil
}
