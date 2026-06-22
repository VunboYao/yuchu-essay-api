package model_test

import (
	"testing"

	"course_server/internal/model"
)

func TestCheckOfficialAccountConfigDeletable(t *testing.T) {
	t.Parallel()

	enabled := model.OfficialAccountConfig{Scene: "article_detail", Status: model.StatusEnabled}
	if err := model.CheckOfficialAccountConfigDeletable(enabled); err == nil {
		t.Fatal("expected enabled official account config not deletable")
	}

	disabled := model.OfficialAccountConfig{Scene: "article_detail", Status: model.StatusDisabled}
	if err := model.CheckOfficialAccountConfigDeletable(disabled); err != nil {
		t.Fatalf("expected disabled official account config deletable, got %v", err)
	}
}
