package admin_test

import (
	"strings"
	"testing"

	admindto "course_server/internal/dto/admin"
	"course_server/internal/model"
)

func TestVideoWriteRequestTrimTitleRequired(t *testing.T) {
	req := admindto.VideoWriteRequest{Title: "   ", FeedID: "export/abc"}
	if _, err := req.ValidateForCreate(); err == nil {
		t.Fatal("expected title is required")
	}
}

func TestVideoWriteRequestGradeOutOfRange(t *testing.T) {
	for _, grade := range []int{7, -1} {
		req := admindto.VideoWriteRequest{Title: "讲解", FeedID: "export/abc", Grade: grade}
		if _, err := req.ValidateForCreate(); err == nil {
			t.Fatalf("expected grade %d rejected", grade)
		}
	}
}

func TestVideoWriteRequestGradeZeroAllowed(t *testing.T) {
	req := admindto.VideoWriteRequest{Title: "通用讲解", FeedID: "export/abc"}
	item, err := req.ValidateForCreate()
	if err != nil {
		t.Fatalf("ValidateForCreate: %v", err)
	}
	if item.Grade != 0 {
		t.Fatalf("grade=%d, want 0", item.Grade)
	}
	if item.Status != model.StatusEnabled {
		t.Fatalf("status=%d, want enabled", item.Status)
	}
}

func TestVideoWriteRequestExplicitGradeZero(t *testing.T) {
	req := admindto.VideoWriteRequest{Title: "通用讲解", FeedID: "export/abc", Grade: 0}
	item, err := req.ValidateForCreate()
	if err != nil {
		t.Fatalf("ValidateForCreate: %v", err)
	}
	if item.Grade != 0 {
		t.Fatalf("grade=%d, want 0", item.Grade)
	}
}

func TestVideoWriteRequestInvalidSemester(t *testing.T) {
	req := admindto.VideoWriteRequest{Title: "讲解", FeedID: "export/abc", Semester: "全册"}
	if _, err := req.ValidateForCreate(); err == nil {
		t.Fatal("expected invalid semester")
	}
}

func TestVideoWriteRequestFeedIDTrimAndMax(t *testing.T) {
	req := admindto.VideoWriteRequest{Title: "讲解", FeedID: "  export/abc  "}
	item, err := req.ValidateForCreate()
	if err != nil {
		t.Fatalf("ValidateForCreate: %v", err)
	}
	if item.FeedID != "export/abc" {
		t.Fatalf("feedId=%q, want trimmed", item.FeedID)
	}

	tooLong := admindto.VideoWriteRequest{Title: "讲解", FeedID: strings.Repeat("a", 256)}
	if _, err := tooLong.ValidateForCreate(); err == nil {
		t.Fatal("expected feedId exceeds max")
	}
}

func TestVideoWriteRequestZeroIDsBecomeNil(t *testing.T) {
	zero := uint(0)
	req := admindto.VideoWriteRequest{
		Title:     "讲解",
		FeedID:    "export/abc",
		ArticleID: &zero,
		EssayID:   &zero,
	}
	item, err := req.ValidateForCreate()
	if err != nil {
		t.Fatalf("ValidateForCreate: %v", err)
	}
	if item.ArticleID != nil || item.EssayID != nil {
		t.Fatalf("expected nil links, got article=%v essay=%v", item.ArticleID, item.EssayID)
	}
}
