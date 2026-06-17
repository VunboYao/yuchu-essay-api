// Package testutil 提供 httptest 用的内存库与路由辅助。
package testutil

import (
	"testing"

	"course_server/internal/database"
	"course_server/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// UseTestDB 用 SQLite 内存库替换全局 database.DB，用例结束自动恢复。
func UseTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	orig := database.DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Article{},
		&model.DownloadFile{},
		&model.AdSlot{},
		&model.DownloadLog{},
		&model.AdEvent{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	t.Cleanup(func() { database.DB = orig })
	return db
}
