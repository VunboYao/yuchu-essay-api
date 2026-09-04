package model_test

import (
	"testing"

	"course_server/internal/model"
)

func TestCheckVideoDeletable(t *testing.T) {
	t.Parallel()

	enabled := model.Video{Title: "上架中", Status: model.StatusEnabled}
	if err := model.CheckVideoDeletable(enabled); err == nil {
		t.Fatal("expected enabled video not deletable")
	}

	disabled := model.Video{Title: "已下架", Status: model.StatusDisabled}
	if err := model.CheckVideoDeletable(disabled); err != nil {
		t.Fatalf("expected disabled video deletable, got %v", err)
	}
}
