package model_test

import (
	"testing"

	"course_server/internal/model"
)

func TestCheckAdSlotDeletable(t *testing.T) {
	t.Parallel()

	enabled := model.AdSlot{SlotKey: "home_feed_mid", Status: model.StatusEnabled}
	if err := model.CheckAdSlotDeletable(enabled); err == nil {
		t.Fatal("expected enabled ad slot not deletable")
	}

	disabled := model.AdSlot{SlotKey: "custom_slot", Status: model.StatusDisabled}
	if err := model.CheckAdSlotDeletable(disabled); err != nil {
		t.Fatalf("expected disabled ad slot deletable, got %v", err)
	}
}
