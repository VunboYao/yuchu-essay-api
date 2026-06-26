package handler_test

import (
	"testing"
	"time"

	"course_server/internal/handler"
	"course_server/internal/model"
	"course_server/internal/pkg/jsontime"
)

func TestBuildDownloadListItemsAvailable(t *testing.T) {
	t.Parallel()
	logs := []model.DownloadLog{{
		ID: 1, FileID: 5, OwnerType: model.TargetTypeArticle, OwnerID: 42,
		AccessType: model.AccessTypeFree, SourceScene: "article_detail",
		CreatedAt: jsontime.Time{Time: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}}
	files := map[uint]model.DownloadFile{
		5: {
			ID: 5, FileName: "资料.pdf", FileType: "pdf", FileSize: 1024,
			Tags: "免费", AccessType: model.AccessTypeFree, Status: model.StatusEnabled,
			OwnerType: model.TargetTypeArticle, OwnerID: 42,
		},
	}
	owners := map[uint]handler.OwnerArticleRef{
		42: {ID: 42, Title: "专题标题", CoverImg: "https://img/cover.jpg"},
	}
	items := handler.BuildDownloadListItems(logs, files, owners)
	if !items[0].Available {
		t.Fatalf("expected available download, got %+v", items[0])
	}
	target, ok := items[0].Target.(handler.DownloadFileSummary)
	if !ok || target.FileName != "资料.pdf" || target.OwnerTitle != "专题标题" || target.CoverImg != "https://img/cover.jpg" {
		t.Fatalf("unexpected target: %+v", items[0].Target)
	}
}

func TestBuildDownloadListItemsDisabledFile(t *testing.T) {
	t.Parallel()
	logs := []model.DownloadLog{{ID: 1, FileID: 5}}
	files := map[uint]model.DownloadFile{
		5: {
			ID: 5, FileName: "旧资料.pdf", FileType: "pdf",
			Tags: "免费", AccessType: model.AccessTypeFree, Status: model.StatusDisabled,
		},
	}
	items := handler.BuildDownloadListItems(logs, files, nil)
	if items[0].Available {
		t.Fatalf("expected unavailable disabled file, got %+v", items[0])
	}
	target, ok := items[0].Target.(handler.DownloadFileSummary)
	if !ok || target.FileName != "旧资料.pdf" {
		t.Fatalf("expected target summary when file exists, got %+v", items[0].Target)
	}
}

func TestBuildDownloadListItemsMissingFile(t *testing.T) {
	t.Parallel()
	logs := []model.DownloadLog{{ID: 1, FileID: 999}}
	items := handler.BuildDownloadListItems(logs, map[uint]model.DownloadFile{}, nil)
	if items[0].Available || items[0].Target != nil {
		t.Fatalf("expected missing file unavailable, got %+v", items[0])
	}
}

func TestBuildDownloadListItemsNotFreeTag(t *testing.T) {
	t.Parallel()
	logs := []model.DownloadLog{{ID: 1, FileID: 5}}
	files := map[uint]model.DownloadFile{
		5: {
			ID: 5, FileName: "资料.pdf", Tags: "付费",
			AccessType: model.AccessTypeFree, Status: model.StatusEnabled,
		},
	}
	items := handler.BuildDownloadListItems(logs, files, nil)
	if items[0].Available {
		t.Fatalf("expected unavailable without free tag, got %+v", items[0])
	}
}

func TestBuildDownloadFileListItemsOwnerTitle(t *testing.T) {
	t.Parallel()
	files := []model.DownloadFile{{
		ID: 1, FileName: "资料.pdf", OwnerType: model.TargetTypeArticle, OwnerID: 42,
	}}
	owners := map[uint]handler.OwnerArticleRef{
		42: {ID: 42, Title: "三年级第一单元"},
	}
	items := handler.BuildDownloadFileListItems(files, owners)
	if items[0].OwnerTitle != "三年级第一单元" {
		t.Fatalf("unexpected owner title: %q", items[0].OwnerTitle)
	}
}
