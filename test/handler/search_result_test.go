package handler_test

import (
	"testing"

	"course_server/internal/handler"
	"course_server/internal/model"
)

// 校验搜索混排：先文章后作文，且字段映射正确。
func TestBuildSearchResultList_ArticleThenEssay(t *testing.T) {
	articles := []model.Article{
		{ID: 1, Title: "专题A", Grade: 3, Semester: "上册", Unit: 1, Summary: "摘要", IsHot: true},
	}
	essays := []model.EssayExample{
		{ID: 9, Title: "范文B", Grade: 4, WordCount: 320},
	}
	list := handler.BuildSearchResultList(articles, essays)
	if len(list) != 2 {
		t.Fatalf("len=%d, want 2", len(list))
	}
	if list[0].TargetType != model.TargetTypeArticle || list[0].ID != 1 || !list[0].IsHot {
		t.Fatalf("article item unexpected: %+v", list[0])
	}
	if list[1].TargetType != model.TargetTypeEssay || list[1].ID != 9 || list[1].WordCount != 320 {
		t.Fatalf("essay item unexpected: %+v", list[1])
	}
}

// 空结果应返回空切片而非 nil，便于前端 list || []。
func TestBuildSearchResultList_Empty(t *testing.T) {
	list := handler.BuildSearchResultList(nil, nil)
	if list == nil || len(list) != 0 {
		t.Fatalf("want empty non-nil slice, got %#v", list)
	}
}
