package handler_test

import (
	"testing"
	"time"

	"course_server/internal/handler"
	"course_server/internal/model"
	"course_server/internal/pkg/jsontime"
)

func TestBuildFavoriteListItemsArticlePublished(t *testing.T) {
	t.Parallel()
	favs := []model.Favorite{{
		ID: 1, TargetType: model.TargetTypeArticle, TargetID: 10,
		CreatedAt: jsontime.Time{Time: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}}
	articles := map[uint]model.Article{
		10: {
			ID: 10, Title: "专题A", CoverImg: "https://img/a.jpg",
			Grade: 3, Semester: "上册", Unit: 1, EssayTopic: "题目", Summary: "摘要",
			Status: model.ArticleStatusPublished,
		},
	}
	items := handler.BuildFavoriteListItems(favs, articles, nil, nil, nil)
	if len(items) != 1 || !items[0].Available {
		t.Fatalf("expected available article item, got %+v", items)
	}
	target, ok := items[0].Target.(handler.FavoriteArticleSummary)
	if !ok || target.Title != "专题A" || target.CoverImg != "https://img/a.jpg" {
		t.Fatalf("unexpected target: %+v", items[0].Target)
	}
}

func TestBuildFavoriteListItemsArticleUnpublished(t *testing.T) {
	t.Parallel()
	favs := []model.Favorite{{ID: 1, TargetType: model.TargetTypeArticle, TargetID: 10}}
	articles := map[uint]model.Article{
		10: {ID: 10, Title: "草稿", Status: model.ArticleStatusDraft},
	}
	items := handler.BuildFavoriteListItems(favs, articles, nil, nil, nil)
	if items[0].Available || items[0].Target != nil {
		t.Fatalf("expected unavailable with nil target, got %+v", items[0])
	}
}

func TestBuildFavoriteListItemsEssayWithLinkedArticle(t *testing.T) {
	t.Parallel()
	articleID := uint(42)
	favs := []model.Favorite{{ID: 2, TargetType: model.TargetTypeEssay, TargetID: 8}}
	essays := map[uint]model.EssayExample{
		8: {ID: 8, Title: "范文B", Grade: 4, WordCount: 500, ArticleID: &articleID, Status: model.StatusEnabled},
	}
	linked := map[uint]handler.ArticleTitleRef{
		42: {ID: 42, Title: "所属专题"},
	}
	items := handler.BuildFavoriteListItems(favs, nil, essays, linked, nil)
	target, ok := items[0].Target.(handler.FavoriteEssaySummary)
	if !ok || !items[0].Available || target.ArticleTitle != "所属专题" || target.ArticleID == nil || *target.ArticleID != 42 {
		t.Fatalf("unexpected essay target: %+v", items[0].Target)
	}
}

func TestBuildFavoriteListItemsMissingTarget(t *testing.T) {
	t.Parallel()
	favs := []model.Favorite{{ID: 3, TargetType: model.TargetTypeEssay, TargetID: 999}}
	items := handler.BuildFavoriteListItems(favs, nil, map[uint]model.EssayExample{}, nil, nil)
	if items[0].Available || items[0].Target != nil {
		t.Fatalf("expected missing essay unavailable, got %+v", items[0])
	}
}

func TestBuildFavoriteListItemsPreservesOrder(t *testing.T) {
	t.Parallel()
	favs := []model.Favorite{
		{ID: 10, TargetType: model.TargetTypeArticle, TargetID: 1},
		{ID: 9, TargetType: model.TargetTypeEssay, TargetID: 2},
	}
	articles := map[uint]model.Article{
		1: {ID: 1, Title: "A", Status: model.ArticleStatusPublished},
	}
	essays := map[uint]model.EssayExample{
		2: {ID: 2, Title: "B", Status: model.StatusEnabled},
	}
	items := handler.BuildFavoriteListItems(favs, articles, essays, nil, nil)
	if len(items) != 2 || items[0].ID != 10 || items[1].ID != 9 {
		t.Fatalf("order mismatch: %+v", items)
	}
}
