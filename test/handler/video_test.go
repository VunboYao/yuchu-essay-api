package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"course_server/internal/handler"
	"course_server/internal/model"
	jwtutil "course_server/internal/pkg/jwt"
	"course_server/test/testutil"

	"gorm.io/gorm"
)

func useVideoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.UseTestDB(t)
	if err := db.AutoMigrate(
		&model.Video{},
		&model.VideoTag{},
		&model.SystemSetting{},
		&model.EssayExample{},
		&model.Tag{},
		&model.EssayExampleTag{},
	); err != nil {
		t.Fatalf("migrate video tables: %v", err)
	}
	return db
}

func adminAuthHeader(t *testing.T) string {
	t.Helper()
	token, err := jwtutil.SignAdmin("test-secret", "essay-mvp-test", 1, "admin", "admin")
	if err != nil {
		t.Fatalf("sign admin: %v", err)
	}
	return "Bearer " + token
}

func postJSON(t *testing.T, r http.Handler, path string, body any, auth string) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func putJSON(t *testing.T, r http.Handler, path string, body any, auth string) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCreateVideoDuplicateFeedID(t *testing.T) {
	db := useVideoDB(t)
	r := testutil.SetupRouter(t)
	if err := db.Create(&model.Video{Title: "已有", FeedID: "export/dup", Status: model.StatusEnabled}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	rec := postJSON(t, r, "/admin/api/v1/videos", map[string]any{
		"title":  "新的",
		"feedId": "export/dup",
	}, adminAuthHeader(t))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := decodeResponseBody(t, rec)
	if body.Message != "feed-id 已存在" {
		t.Fatalf("message=%q", body.Message)
	}
}

func TestReplaceVideoTagsNilKeepsExisting(t *testing.T) {
	db := useVideoDB(t)
	r := testutil.SetupRouter(t)
	auth := adminAuthHeader(t)

	tag := model.Tag{Name: "习作", Status: model.StatusEnabled}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("tag: %v", err)
	}
	rec := postJSON(t, r, "/admin/api/v1/videos", map[string]any{
		"title":  "带标签",
		"feedId": "export/keep",
		"tagIds": []uint{tag.ID},
	}, auth)
	if rec.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", rec.Code, rec.Body.String())
	}

	var created model.Video
	if err := db.Where("feed_id = ?", "export/keep").First(&created).Error; err != nil {
		t.Fatalf("load video: %v", err)
	}
	rec = putJSON(t, r, "/admin/api/v1/videos/"+strconv.FormatUint(uint64(created.ID), 10), map[string]any{
		"title":  "带标签改名",
		"feedId": "export/keep",
	}, auth)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", rec.Code, rec.Body.String())
	}
	var count int64
	if err := db.Model(&model.VideoTag{}).Where("video_id = ?", created.ID).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("tag count=%d, want 1 (nil tagIds should keep)", count)
	}
}

func TestReplaceVideoTagsEmptyClears(t *testing.T) {
	db := useVideoDB(t)
	r := testutil.SetupRouter(t)
	auth := adminAuthHeader(t)

	tag := model.Tag{Name: "清空", Status: model.StatusEnabled}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("tag: %v", err)
	}
	rec := postJSON(t, r, "/admin/api/v1/videos", map[string]any{
		"title":  "将清空",
		"feedId": "export/clear",
		"tagIds": []uint{tag.ID},
	}, auth)
	if rec.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", rec.Code, rec.Body.String())
	}
	var created model.Video
	if err := db.Where("feed_id = ?", "export/clear").First(&created).Error; err != nil {
		t.Fatalf("load video: %v", err)
	}
	rec = putJSON(t, r, "/admin/api/v1/videos/"+strconv.FormatUint(uint64(created.ID), 10), map[string]any{
		"title":  "将清空",
		"feedId": "export/clear",
		"tagIds": []uint{},
	}, auth)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", rec.Code, rec.Body.String())
	}
	var count int64
	if err := db.Model(&model.VideoTag{}).Where("video_id = ?", created.ID).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("tag count=%d, want 0", count)
	}
}

func TestListPublicVideosGradeZeroVisibleAndFinderTopLevel(t *testing.T) {
	db := useVideoDB(t)
	r := testutil.SetupRouter(t)
	if err := db.Create(&model.SystemSetting{
		SettingKey:   model.SettingKeyChannelsFinderUserName,
		SettingValue: "sph123",
	}).Error; err != nil {
		t.Fatalf("setting: %v", err)
	}
	if err := db.Create([]model.Video{
		{Title: "不限年级", FeedID: "export/g0", Grade: 0, Status: model.StatusEnabled},
		{Title: "三年级", FeedID: "export/g3", Grade: 3, Status: model.StatusEnabled},
	}).Error; err != nil {
		t.Fatalf("seed videos: %v", err)
	}
	off := model.Video{Title: "已下架", FeedID: "export/off", Grade: 0}
	if err := db.Create(&off).Error; err != nil {
		t.Fatalf("seed off: %v", err)
	}
	if err := db.Model(&off).Update("status", model.StatusDisabled).Error; err != nil {
		t.Fatalf("disable video: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/videos?grade=3", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := decodeResponseBody(t, rec)
	data, _ := body.Data.(map[string]any)
	if data["finderUserName"] != "sph123" {
		t.Fatalf("finderUserName=%v", data["finderUserName"])
	}
	list, _ := data["list"].([]any)
	titles := map[string]bool{}
	for _, row := range list {
		item, _ := row.(map[string]any)
		title, _ := item["title"].(string)
		titles[title] = true
		if _, ok := item["finderUserName"]; ok {
			t.Fatal("finderUserName must not repeat on each item")
		}
	}
	if !titles["不限年级"] || !titles["三年级"] || titles["已下架"] {
		t.Fatalf("unexpected titles: %v", titles)
	}
}

func TestBuildPublicVideoItemsUnpublishedArticleIsNull(t *testing.T) {
	articleID := uint(10)
	essayID := uint(20)
	rows := []model.Video{{
		ID: 1, Title: "讲解", FeedID: "export/x",
		ArticleID: &articleID, EssayID: &essayID,
	}}
	articles := map[uint]model.Article{
		10: {ID: 10, Title: "草稿专题", Status: model.ArticleStatusDraft, CoverImg: "covers/a.jpg"},
	}
	essays := map[uint]model.EssayExample{
		20: {ID: 20, Title: "范文", Status: model.StatusEnabled},
	}
	items := handler.BuildPublicVideoItems(rows, articles, essays, nil, func(s string) string {
		return "https://cdn.example/" + s
	})
	if len(items) != 1 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0].Article != nil {
		t.Fatalf("unpublished article should be null, got %+v", items[0].Article)
	}
	if items[0].Essay == nil || items[0].Essay.ID != 20 {
		t.Fatalf("expected enabled essay, got %+v", items[0].Essay)
	}
}
