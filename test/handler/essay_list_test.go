package handler_test

import (
	"testing"

	"course_server/internal/handler"
	"course_server/internal/model"
)

func TestBuildEssayListItemsWithTags(t *testing.T) {
	t.Parallel()
	essays := []model.EssayExample{
		{ID: 1, Title: "范文A"},
		{ID: 2, Title: "范文B"},
	}
	tagsByEssayID := map[uint][]model.Tag{
		1: {{ID: 10, Name: "记叙文"}, {ID: 11, Name: "三年级"}},
	}
	items := handler.BuildEssayListItems(essays, tagsByEssayID)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "范文A" || len(items[0].Tags) != 2 {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[0].Tags[0].Name != "记叙文" || items[0].Tags[1].Name != "三年级" {
		t.Fatalf("unexpected tags order: %+v", items[0].Tags)
	}
	if len(items[1].Tags) != 0 {
		t.Fatalf("expected empty tags for essay without mapping, got %+v", items[1].Tags)
	}
}

func TestBuildEssayListItemsEmpty(t *testing.T) {
	t.Parallel()
	items := handler.BuildEssayListItems(nil, nil)
	if items == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(items) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(items))
	}
}
