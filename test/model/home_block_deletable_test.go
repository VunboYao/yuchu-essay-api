package model_test

import (
	"testing"

	"course_server/internal/model"
)

func TestCheckHomeBlockDeletable(t *testing.T) {
	t.Parallel()

	enabled := model.HomeBlock{Title: "写作素材库", Status: model.StatusEnabled}
	if err := model.CheckHomeBlockDeletable(enabled); err == nil {
		t.Fatal("expected enabled home block not deletable")
	}

	disabled := model.HomeBlock{Title: "已下线入口", Status: model.StatusDisabled}
	if err := model.CheckHomeBlockDeletable(disabled); err != nil {
		t.Fatalf("expected disabled home block deletable, got %v", err)
	}
}
