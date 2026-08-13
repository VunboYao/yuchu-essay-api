package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"course_server/internal/model"
	"course_server/internal/pkg/adunlock"
	"course_server/internal/pkg/response"
	"course_server/test/testutil"

	"gorm.io/gorm"
)

func decodeResponseBody(t *testing.T, rec *httptest.ResponseRecorder) response.Body {
	t.Helper()
	var body response.Body
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, rec.Body.String())
	}
	return body
}

func seedFreeDownloadFile(t *testing.T, db *gorm.DB, overrides func(*model.DownloadFile)) model.DownloadFile {
	t.Helper()
	file := model.DownloadFile{
		OwnerType:  model.TargetTypeArticle,
		OwnerID:    1,
		FileName:   "测试资料.pdf",
		Tags:       "免费",
		AccessType: model.AccessTypeFree,
		FileType:   "pdf",
		FileURL:    "https://oss.example.com/test.pdf",
		Status:     model.StatusEnabled,
		NeedLogin:  false,
	}
	if overrides != nil {
		overrides(&file)
	}
	needLogin := file.NeedLogin
	status := file.Status
	if err := db.Create(&file).Error; err != nil {
		t.Fatalf("seed file: %v", err)
	}
	// Create 会把 default:true 回写到 NeedLogin，须用创建前的值强制落库
	if err := db.Model(&file).Updates(map[string]interface{}{
		"need_login": needLogin,
		"status":     status,
	}).Error; err != nil {
		t.Fatalf("seed file flags: %v", err)
	}
	file.NeedLogin = needLogin
	file.Status = status
	return file
}

func signURLPath(fileID uint) string {
	return "/api/v1/download-files/" + strconv.FormatUint(uint64(fileID), 10) + "/sign-url"
}

func TestSignDownloadURLNotFound(t *testing.T) {
	testutil.UseTestDB(t)
	r := testutil.SetupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/download-files/999/sign-url", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestSignDownloadURLForbiddenDisabled(t *testing.T) {
	db := testutil.UseTestDB(t)
	r := testutil.SetupRouter(t)
	file := seedFreeDownloadFile(t, db, nil)
	if err := db.Model(&file).UpdateColumn("status", model.StatusDisabled).Error; err != nil {
		t.Fatalf("disable file: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, signURLPath(file.ID), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}
	body := decodeResponseBody(t, rec)
	if body.Message != "download file is not free" {
		t.Fatalf("message=%q", body.Message)
	}
}

func TestSignDownloadURLForbiddenPurchaseRequired(t *testing.T) {
	db := testutil.UseTestDB(t)
	r := testutil.SetupRouter(t)
	file := seedFreeDownloadFile(t, db, func(f *model.DownloadFile) {
		f.AccessType = model.AccessTypePurchaseRequired
	})

	req := httptest.NewRequest(http.MethodGet, signURLPath(file.ID), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestSignDownloadURLForbiddenGateWithoutToken(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, signURLPath(file.ID), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}
	body := decodeResponseBody(t, rec)
	if body.Message != "rewarded video required" {
		t.Fatalf("message=%q", body.Message)
	}
}

func TestSignDownloadURLSuccessWithoutGate(t *testing.T) {
	db := testutil.UseTestDB(t)
	r := testutil.SetupRouter(t)
	file := seedFreeDownloadFile(t, db, nil)

	req := httptest.NewRequest(http.MethodGet, signURLPath(file.ID), nil)
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

func TestSignDownloadURLSuccessWithValidUnlockToken(t *testing.T) {
	db := testutil.UseTestDB(t)
	r := testutil.SetupRouter(t)
	cfg := testutil.TestConfig()
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
	token, _, err := adunlock.Sign(cfg.JWT.Secret, cfg.JWT.Issuer, file.ID, 0, model.AdSlotKeyDownloadRewardedVideo)
	if err != nil {
		t.Fatalf("sign unlock: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, signURLPath(file.ID)+"?unlockToken="+token, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}
}
