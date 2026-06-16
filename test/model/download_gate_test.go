package model_test

import (
	"testing"

	"course_server/internal/model"
)

func TestBuildDownloadGateInactive(t *testing.T) {
	gate := model.BuildDownloadGate(model.AdSlot{
		SlotKey:  model.AdSlotKeyDownloadRewardedVideo,
		AdType:   model.AdTypeRewardedVideo,
		Status:   model.StatusDisabled,
		AdUnitID: "adunit-test",
	})
	if gate.Required {
		t.Fatal("expected required=false when disabled")
	}
}

func TestBuildDownloadGateActive(t *testing.T) {
	slot := model.AdSlot{
		SlotKey:  model.AdSlotKeyDownloadRewardedVideo,
		AdType:   model.AdTypeRewardedVideo,
		Page:     model.AdPageDownload,
		Status:   model.StatusEnabled,
		AdUnitID: "adunit-test",
	}
	if !slot.IsDownloadGateActive() {
		t.Fatal("expected gate active")
	}
	gate := model.BuildDownloadGate(slot)
	if !gate.Required || gate.AdType != model.AdTypeRewardedVideo || gate.AdUnitID != "adunit-test" {
		t.Fatalf("unexpected gate: %+v", gate)
	}
}

func TestBuildDownloadGateMissingUnitID(t *testing.T) {
	gate := model.BuildDownloadGate(model.AdSlot{
		SlotKey: model.AdSlotKeyDownloadRewardedVideo,
		AdType:  model.AdTypeRewardedVideo,
		Status:  model.StatusEnabled,
	})
	if gate.Required {
		t.Fatal("expected required=false without adUnitId")
	}
}
