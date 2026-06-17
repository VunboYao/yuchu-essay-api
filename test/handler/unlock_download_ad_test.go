package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"course_server/internal/model"
	"course_server/test/testutil"
)

func adUnlockPath(fileID uint) string {
	return "/api/v1/download-files/" + strconv.FormatUint(uint64(fileID), 10) + "/ad-unlock"
}

func TestUnlockDownloadAdForbiddenNotFree(t *testing.T) {
	db := testutil.UseTestDB(t)
	r := testutil.SetupRouter(t)
	file := seedFreeDownloadFile(t, db, func(f *model.DownloadFile) {
		f.Tags = "PDF"
	})

	req := httptest.NewRequest(http.MethodPost, adUnlockPath(file.ID), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestUnlockDownloadAdBadRequestGateDisabled(t *testing.T) {
	db := testutil.UseTestDB(t)
	r := testutil.SetupRouter(t)
	file := seedFreeDownloadFile(t, db, nil)

	req := httptest.NewRequest(http.MethodPost, adUnlockPath(file.ID), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}
	body := decodeResponseBody(t, rec)
	if body.Message != "download rewarded video gate is not enabled" {
		t.Fatalf("message=%q", body.Message)
	}
}

func TestUnlockDownloadAdSuccess(t *testing.T) {
	db := testutil.UseTestDB(t)
	r := testutil.SetupRouter(t)
	file := seedFreeDownloadFile(t, db, nil)
	if err := db.Create(&model.AdSlot{
		SlotKey:  model.AdSlotKeyDownloadRewardedVideo,
		SlotName: "下载激励视频",
		Page:     model.AdPageDownload,
		AdType:   model.AdTypeRewardedVideo,
		AdUnitID: "adunit-test",
		Status:   model.StatusEnabled,
	}).Error; err != nil {
		t.Fatalf("seed gate: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, adUnlockPath(file.ID), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}
	body := decodeResponseBody(t, rec)
	if body.Code != 0 {
		t.Fatalf("code=%d message=%s", body.Code, body.Message)
	}
}
