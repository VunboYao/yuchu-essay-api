package handler_test

import (
	"testing"
	"time"

	"course_server/internal/handler"
	"course_server/internal/model"
	"course_server/internal/pkg/jsontime"
)

func TestBuildHistoryListItemsArticle(t *testing.T) {
	t.Parallel()
	logs := []model.ViewLog{{
		ID: 1, TargetType: model.TargetTypeArticle, TargetID: 10,
		PagePath: "pages/article/detail", SourceScene: "home",
		CreatedAt: jsontime.Time{Time: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}}
	articles := map[uint]model.Article{
		10: {ID: 10, Title: "专题", CoverImg: "https://img/a.jpg", Status: model.ArticleStatusPublished},
	}
	items := handler.BuildHistoryListItems(logs, articles, nil, nil)
	if !items[0].Available || items[0].PagePath != "pages/article/detail" {
		t.Fatalf("unexpected item: %+v", items[0])
	}
}

func TestBuildHistoryListItemsUnavailable(t *testing.T) {
	t.Parallel()
	logs := []model.ViewLog{{ID: 1, TargetType: model.TargetTypeArticle, TargetID: 10}}
	articles := map[uint]model.Article{
		10: {ID: 10, Title: "草稿", Status: model.ArticleStatusDraft},
	}
	items := handler.BuildHistoryListItems(logs, articles, nil, nil)
	if items[0].Available || items[0].Target != nil {
		t.Fatalf("expected unavailable, got %+v", items[0])
	}
}

func TestBuildHistoryListItemsUnknownType(t *testing.T) {
	t.Parallel()
	logs := []model.ViewLog{{ID: 1, TargetType: "unknown", TargetID: 1}}
	items := handler.BuildHistoryListItems(logs, nil, nil, nil)
	if items[0].Available || items[0].Target != nil {
		t.Fatalf("expected unknown type unavailable, got %+v", items[0])
	}
}

func TestBuildHistoryListItemsPreservesOrder(t *testing.T) {
	t.Parallel()
	logs := []model.ViewLog{
		{ID: 2, TargetType: model.TargetTypeEssay, TargetID: 1},
		{ID: 1, TargetType: model.TargetTypeArticle, TargetID: 2},
	}
	essays := map[uint]model.EssayExample{
		1: {ID: 1, Title: "E", Status: model.StatusEnabled},
	}
	articles := map[uint]model.Article{
		2: {ID: 2, Title: "A", Status: model.ArticleStatusPublished},
	}
	items := handler.BuildHistoryListItems(logs, articles, essays, nil)
	if items[0].ID != 2 || items[1].ID != 1 {
		t.Fatalf("order mismatch: %+v", items)
	}
}
