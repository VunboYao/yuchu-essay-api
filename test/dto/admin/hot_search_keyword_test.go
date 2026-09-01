package admin_test

import (
	"strings"
	"testing"

	admindto "course_server/internal/dto/admin"
	"course_server/internal/model"
)

func TestHotSearchKeywordWriteRequestTrimAndRequired(t *testing.T) {
	req := admindto.HotSearchKeywordWriteRequest{Keyword: "   "}
	if _, err := req.ValidateForCreate(); err == nil {
		t.Fatal("expected keyword is required")
	}
}

func TestHotSearchKeywordWriteRequestMaxRunes(t *testing.T) {
	req := admindto.HotSearchKeywordWriteRequest{Keyword: strings.Repeat("词", 41)}
	if _, err := req.ValidateForCreate(); err == nil {
		t.Fatal("expected keyword exceeds max 40")
	}
}

func TestHotSearchKeywordWriteRequestCreateDefaultsEnabled(t *testing.T) {
	req := admindto.HotSearchKeywordWriteRequest{Keyword: "  写人作文  "}
	item, err := req.ValidateForCreate()
	if err != nil {
		t.Fatalf("ValidateForCreate: %v", err)
	}
	if item.Keyword != "写人作文" {
		t.Fatalf("keyword = %q, want trimmed", item.Keyword)
	}
	if item.Status != model.StatusEnabled {
		t.Fatalf("status = %d, want enabled", item.Status)
	}
}
