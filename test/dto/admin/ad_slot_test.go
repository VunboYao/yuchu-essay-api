package admin_test

import (
	"testing"

	admindto "course_server/internal/dto/admin"
	"course_server/internal/model"
)

func TestUpsertAdSlotRequestInvalidAdType(t *testing.T) {
	req := admindto.UpsertAdSlotRequest{
		SlotKey:  "x",
		SlotName: "x",
		Page:     model.AdPageHome,
		AdType:   "interstitial",
	}
	if _, err := req.Validate(model.StatusEnabled); err == nil {
		t.Fatal("expected invalid adType error")
	}
}

func TestUpsertAdSlotRequestRewardedVideoPage(t *testing.T) {
	req := admindto.UpsertAdSlotRequest{
		SlotKey:  model.AdSlotKeyDownloadRewardedVideo,
		SlotName: "下载激励",
		Page:     model.AdPageHome,
		AdType:   model.AdTypeRewardedVideo,
	}
	if _, err := req.Validate(model.StatusEnabled); err == nil {
		t.Fatal("expected page must be download error")
	}
}

func TestUpsertAdSlotRequestDefaultsBanner(t *testing.T) {
	req := admindto.UpsertAdSlotRequest{
		SlotKey:  model.AdSlotKeyHomeFeedMid,
		SlotName: "首页",
		Page:     model.AdPageHome,
	}
	slot, err := req.Validate(model.StatusEnabled)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if slot.AdType != model.AdTypeBanner {
		t.Fatalf("adType = %q, want banner", slot.AdType)
	}
}
