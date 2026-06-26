package handler_test

import (
	"testing"

	"course_server/internal/handler"
	"course_server/internal/model"
)

func TestBuildArticleListItemsWithTags(t *testing.T) {
	t.Parallel()
	articles := []model.Article{
		{ID: 1, Title: "专题A"},
		{ID: 2, Title: "专题B"},
	}
	tagsByArticleID := map[uint][]model.Tag{
		1: {{ID: 10, Name: "三年级"}, {ID: 11, Name: "单元习作"}},
	}
	items := handler.BuildArticleListItems(articles, tagsByArticleID, nil)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "专题A" || len(items[0].Tags) != 2 {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[0].Tags[0].Name != "三年级" || items[0].Tags[1].Name != "单元习作" {
		t.Fatalf("unexpected tags order: %+v", items[0].Tags)
	}
	if len(items[1].Tags) != 0 {
		t.Fatalf("expected empty tags for article without mapping, got %+v", items[1].Tags)
	}
	if items[0].DownloadSummary.Total != 0 || items[0].DownloadSummary.PublicFree != 0 {
		t.Fatalf("expected zero download summary when map nil, got %+v", items[0].DownloadSummary)
	}
}

func TestBuildArticleListItemsWithDownloadSummary(t *testing.T) {
	t.Parallel()
	articles := []model.Article{{ID: 1, Title: "专题A"}}
	downloads := map[uint]handler.ArticleDownloadSummary{
		1: {Total: 2, PublicFree: 1},
	}
	items := handler.BuildArticleListItems(articles, nil, downloads)
	if items[0].DownloadSummary.Total != 2 || items[0].DownloadSummary.PublicFree != 1 {
		t.Fatalf("unexpected download summary: %+v", items[0].DownloadSummary)
	}
}

func TestBuildArticleListItemsEmpty(t *testing.T) {
	t.Parallel()
	items := handler.BuildArticleListItems(nil, nil, nil)
	if items == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(items) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(items))
	}
}
