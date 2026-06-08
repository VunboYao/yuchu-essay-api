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
	Slug            string   `json:"slug" binding:"required"`
	TextbookVersion string   `json:"textbookVersion"`
	Grade           int      `json:"grade" binding:"required"`
	Semester        string   `json:"semester" binding:"required"`
	Unit            int      `json:"unit" binding:"required"`
	EssayTopic      string   `json:"essayTopic" binding:"required"`
	ContentType     string   `json:"contentType" binding:"required"`
	Audience        string   `json:"audience"`
	Summary         string   `json:"summary" binding:"required"`
	CoverImg        string   `json:"coverImg" binding:"required"`
	BodyImages      []string `json:"bodyImages"`
	ViewAccessType  string   `json:"viewAccessType"`
	Status          *int     `json:"status"`
	IsRecommended   bool     `json:"isRecommended"`
	IsHot           bool     `json:"isHot"`
	SortOrder       int      `json:"sortOrder"`
	TagIDs          []uint   `json:"tagIds"`
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

// 文章参数校验与转换；textbookVersion 默认值由 ValidateForCreate/ValidateForUpdate 处理。
func (r ArticleWriteRequest) normalize() (model.Article, error) {
	// 校验必填字符串字段
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return model.Article{}, fmt.Errorf("title is required")
	}
	slug := strings.TrimSpace(r.Slug)
	if slug == "" {
		return model.Article{}, fmt.Errorf("slug is required")
	}
	version := strings.TrimSpace(r.TextbookVersion)
	// 校验必填数值字段：年级和单元须为正整数
	if r.Grade <= 0 {
		return model.Article{}, fmt.Errorf("grade must be greater than 0")
	}
	semester := strings.TrimSpace(r.Semester)
	if semester == "" {
		return model.Article{}, fmt.Errorf("semester is required")
	}
	if !model.ValidSemester(semester) {
		return model.Article{}, fmt.Errorf("invalid semester")
	}
	if r.Unit <= 0 {
		return model.Article{}, fmt.Errorf("unit must be greater than 0")
	}
	essayTopic := strings.TrimSpace(r.EssayTopic)
	if essayTopic == "" {
		return model.Article{}, fmt.Errorf("essayTopic is required")
	}
	contentType := strings.TrimSpace(r.ContentType)
	if contentType == "" {
		return model.Article{}, fmt.Errorf("contentType is required")
	}
	audience := strings.TrimSpace(r.Audience)
	summary := strings.TrimSpace(r.Summary)
	if summary == "" {
		return model.Article{}, fmt.Errorf("summary is required")
	}
	coverImg := strings.TrimSpace(r.CoverImg)
	if coverImg == "" {
		return model.Article{}, fmt.Errorf("coverImg is required")
	}
	// bodyImages 非必填，但若传入则校验内容合法性
	if err := model.ValidateBodyImages(r.BodyImages); err != nil {
		return model.Article{}, err
	}
	// viewAccessType 数据库有默认值 free，未传则使用默认
	viewAccess := strings.TrimSpace(r.ViewAccessType)
	if viewAccess == "" {
		viewAccess = model.ViewAccessFree
	}
	if !model.ValidViewAccessType(viewAccess) {
		return model.Article{}, fmt.Errorf("invalid viewAccessType")
	}
	return model.Article{
		Title:           title,
		Slug:            slug,
		TextbookVersion: version,
		Grade:           r.Grade,
		Semester:        semester,
		Unit:            r.Unit,
		EssayTopic:      essayTopic,
		ContentType:     contentType,
		Audience:        audience,
		Summary:         summary,
		CoverImg:        coverImg,
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
	if article.TextbookVersion == "" {
		article.TextbookVersion = model.DefaultTextbookVersion
	}
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
	if article.TextbookVersion == "" {
		article.TextbookVersion = strings.TrimSpace(existing.TextbookVersion)
		if article.TextbookVersion == "" {
			article.TextbookVersion = model.DefaultTextbookVersion
		}
	}
	return article, nil
}
