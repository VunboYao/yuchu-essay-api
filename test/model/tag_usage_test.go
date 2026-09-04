package model_test

import (
	"strings"
	"testing"

	"course_server/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTagTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Tag{}, &model.ArticleTag{}, &model.EssayExampleTag{}, &model.VideoTag{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestCheckTagDisableAllowed(t *testing.T) {
	t.Parallel()
	db := openTagTestDB(t)

	tag := model.Tag{Name: "写景", Status: model.StatusEnabled}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if err := model.CheckTagDisableAllowed(db, tag.ID); err != nil {
		t.Fatalf("expected no usage error, got %v", err)
	}

	if err := db.Create(&model.ArticleTag{ArticleID: 1, TagID: tag.ID}).Error; err != nil {
		t.Fatalf("create article tag: %v", err)
	}
	if err := model.CheckTagDisableAllowed(db, tag.ID); err == nil {
		t.Fatal("expected disable blocked when article linked")
	}
}

func TestCheckTagDeletable(t *testing.T) {
	t.Parallel()
	db := openTagTestDB(t)

	enabled := model.Tag{Name: "启用", Status: model.StatusEnabled}
	if err := db.Create(&enabled).Error; err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if err := model.CheckTagDeletable(db, enabled); err == nil {
		t.Fatal("expected enabled tag not deletable")
	}

	disabled := model.Tag{Name: "禁用", Status: model.StatusDisabled}
	if err := db.Create(&disabled).Error; err != nil {
		t.Fatalf("create tag: %v", err)
	}
	// GORM default:1 可能覆盖零值 status，以库中记录为准
	if err := db.First(&disabled, disabled.ID).Error; err != nil {
		t.Fatalf("reload tag: %v", err)
	}
	if err := db.Model(&disabled).Update("status", model.StatusDisabled).Error; err != nil {
		t.Fatalf("set disabled: %v", err)
	}
	if err := db.First(&disabled, disabled.ID).Error; err != nil {
		t.Fatalf("reload tag: %v", err)
	}
	if err := model.CheckTagDeletable(db, disabled); err != nil {
		t.Fatalf("expected disabled unused tag deletable, got %v", err)
	}

	if err := db.Create(&model.EssayExampleTag{EssayID: 2, TagID: disabled.ID}).Error; err != nil {
		t.Fatalf("create essay tag: %v", err)
	}
	if err := model.CheckTagDeletable(db, disabled); err == nil {
		t.Fatal("expected delete blocked when essay linked")
	}
}

func TestCheckTagDisableAllowedBlockedByVideo(t *testing.T) {
	t.Parallel()
	db := openTagTestDB(t)

	tag := model.Tag{Name: "讲解", Status: model.StatusEnabled}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if err := db.Create(&model.VideoTag{VideoID: 1, TagID: tag.ID}).Error; err != nil {
		t.Fatalf("create video tag: %v", err)
	}
	err := model.CheckTagDisableAllowed(db, tag.ID)
	if err == nil {
		t.Fatal("expected disable blocked when video linked")
	}
	if !strings.Contains(err.Error(), "讲解视频") {
		t.Fatalf("error should mention 讲解视频, got %v", err)
	}
}
