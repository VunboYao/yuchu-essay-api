package model_test

import (
	"testing"

	"course_server/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTaxonomyUnitTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.TaxonomyUnit{}, &model.Article{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func createDisabledTaxonomyUnit(t *testing.T, db *gorm.DB) model.TaxonomyUnit {
	t.Helper()
	unit := model.TaxonomyUnit{
		TextbookVersion: model.DefaultTextbookVersion,
		Grade:           3,
		Semester:        "上册",
		Unit:            1,
		UnitTitle:       "第一单元",
		Status:          model.StatusDisabled,
	}
	if err := db.Create(&unit).Error; err != nil {
		t.Fatalf("create taxonomy unit: %v", err)
	}
	// GORM default:1 可能覆盖零值 status，以库中记录为准
	if err := db.First(&unit, unit.ID).Error; err != nil {
		t.Fatalf("reload taxonomy unit: %v", err)
	}
	if err := db.Model(&unit).Update("status", model.StatusDisabled).Error; err != nil {
		t.Fatalf("set disabled: %v", err)
	}
	if err := db.First(&unit, unit.ID).Error; err != nil {
		t.Fatalf("reload taxonomy unit: %v", err)
	}
	return unit
}

func TestCheckTaxonomyUnitDeletable(t *testing.T) {
	t.Parallel()
	db := openTaxonomyUnitTestDB(t)

	enabled := model.TaxonomyUnit{
		TextbookVersion: model.DefaultTextbookVersion,
		Grade:           1,
		Semester:        "上册",
		Unit:            1,
		Status:          model.StatusEnabled,
	}
	if err := db.Create(&enabled).Error; err != nil {
		t.Fatalf("create enabled unit: %v", err)
	}
	if err := model.CheckTaxonomyUnitDeletable(db, enabled); err == nil {
		t.Fatal("expected enabled taxonomy unit not deletable")
	}

	disabled := createDisabledTaxonomyUnit(t, db)
	if err := model.CheckTaxonomyUnitDeletable(db, disabled); err != nil {
		t.Fatalf("expected disabled unused unit deletable, got %v", err)
	}

	article := model.Article{
		Title:           "占用单元的文章",
		TextbookVersion: disabled.TextbookVersion,
		Grade:           disabled.Grade,
		Semester:        disabled.Semester,
		Unit:            disabled.Unit,
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("create article: %v", err)
	}
	if err := model.CheckTaxonomyUnitDeletable(db, disabled); err == nil {
		t.Fatal("expected delete blocked when article uses the same tuple")
	}
}
