// 作文专题 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package admin

import (
	"fmt"
	"strings"

	"course_server/internal/model"
)

// ArticleWriteRequest 文章创建/更新共用入参；不含 id、时间戳、计数与 publishedAt。
type ArticleWriteRequest struct {
	Title           string   `json:"title" binding:"required"`
	Slug            string   `json:"slug"`
	TextbookVersion string   `json:"textbookVersion"`
	Grade           int      `json:"grade"`
	Semester        string   `json:"semester"`
	Unit            int      `json:"unit"`
	EssayTopic      string   `json:"essayTopic"`
	ContentType     string   `json:"contentType"`
	Audience        string   `json:"audience"`
	Summary         string   `json:"summary"`
	CoverImg        string   `json:"coverImg"`
	BodyImages      []string `json:"bodyImages"`
	ViewAccessType  string   `json:"viewAccessType"`
	Status          *int     `json:"status"`
	IsRecommended   bool     `json:"isRecommended"`
	IsHot           bool     `json:"isHot"`
	SortOrder       int      `json:"sortOrder"`
}

// CreateArticleRequest 创建文章请求。
type CreateArticleRequest = ArticleWriteRequest

// UpdateArticleRequest 更新文章请求。
type UpdateArticleRequest = ArticleWriteRequest

func resolveArticleStatus(ptr *int, defaultStatus int) (int, error) {
	if ptr == nil {
		return defaultStatus, nil
	}
	if *ptr != model.ArticleStatusDraft && *ptr != model.ArticleStatusPublished {
		return 0, fmt.Errorf("invalid status")
	}
	return *ptr, nil
}

func (r ArticleWriteRequest) normalize() (model.Article, error) {
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return model.Article{}, fmt.Errorf("title is required")
	}
	if err := model.ValidateBodyImages(r.BodyImages); err != nil {
		return model.Article{}, err
	}
	version := strings.TrimSpace(r.TextbookVersion)
	if version == "" {
		version = model.DefaultTextbookVersion
	}
	semester := strings.TrimSpace(r.Semester)
	if semester != "" && !model.ValidSemester(semester) {
		return model.Article{}, fmt.Errorf("invalid semester")
	}
	viewAccess := strings.TrimSpace(r.ViewAccessType)
	if viewAccess == "" {
		viewAccess = model.ViewAccessFree
	}
	if !model.ValidViewAccessType(viewAccess) {
		return model.Article{}, fmt.Errorf("invalid viewAccessType")
	}
	return model.Article{
		Title:           title,
		Slug:            strings.TrimSpace(r.Slug),
		TextbookVersion: version,
		Grade:           r.Grade,
		Semester:        semester,
		Unit:            r.Unit,
		EssayTopic:      strings.TrimSpace(r.EssayTopic),
		ContentType:     strings.TrimSpace(r.ContentType),
		Audience:        strings.TrimSpace(r.Audience),
		Summary:         strings.TrimSpace(r.Summary),
		CoverImg:        strings.TrimSpace(r.CoverImg),
		BodyImages:      r.BodyImages,
		ViewAccessType:  viewAccess,
		IsRecommended:   r.IsRecommended,
		IsHot:           r.IsHot,
		SortOrder:       r.SortOrder,
	}, nil
}

// ValidateForCreate 校验创建入参；status 未传时默认草稿。
func (r ArticleWriteRequest) ValidateForCreate() (model.Article, error) {
	article, err := r.normalize()
	if err != nil {
		return model.Article{}, err
	}
	status, err := resolveArticleStatus(r.Status, model.ArticleStatusDraft)
	if err != nil {
		return model.Article{}, err
	}
	article.Status = status
	return article, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r ArticleWriteRequest) ValidateForUpdate(existing model.Article) (model.Article, error) {
	article, err := r.normalize()
	if err != nil {
		return model.Article{}, err
	}
	status, err := resolveArticleStatus(r.Status, existing.Status)
	if err != nil {
		return model.Article{}, err
	}
	article.Status = status
	return article, nil
}
