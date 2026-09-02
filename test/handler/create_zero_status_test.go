package handler_test

import (
	"testing"

	"course_server/internal/handler"
	"course_server/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateAndReload_persistsDisabledStatus(t *testing.T) {
	t.Parallel()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Tag{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	want := model.Tag{Name: "禁用新建", SortOrder: 1, Status: model.StatusDisabled}
	if err := handler.CreateAndReload(db, &want, want.UpdateMap()); err != nil {
		t.Fatalf("CreateAndReload: %v", err)
	}
	if want.ID == 0 {
		t.Fatal("expected id after create")
	}
	if want.Status != model.StatusDisabled {
		t.Fatalf("CreateAndReload status=%d, want 0", want.Status)
	}

	skipped := model.Tag{Name: "对照", Status: model.StatusDisabled}
	if err := db.Create(&skipped).Error; err != nil {
		t.Fatalf("raw Create: %v", err)
	}
	if err := db.First(&skipped, skipped.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if skipped.Status != model.StatusEnabled {
		t.Fatalf("raw Create status=%d, want 1 (GORM skipped zero)", skipped.Status)
	}
}
