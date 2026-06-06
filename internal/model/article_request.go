// 作文专题 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
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
	if *ptr != ArticleStatusDraft && *ptr != ArticleStatusPublished {
		return 0, fmt.Errorf("invalid status")
	}
	return *ptr, nil
}

func (r ArticleWriteRequest) normalize() (Article, error) {
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return Article{}, fmt.Errorf("title is required")
	}
	if err := ValidateBodyImages(r.BodyImages); err != nil {
		return Article{}, err
	}
	version := strings.TrimSpace(r.TextbookVersion)
	if version == "" {
		version = DefaultTextbookVersion
	}
	semester := strings.TrimSpace(r.Semester)
	if semester != "" && !ValidSemester(semester) {
		return Article{}, fmt.Errorf("invalid semester")
	}
	return Article{
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
		IsRecommended:   r.IsRecommended,
		IsHot:           r.IsHot,
		SortOrder:       r.SortOrder,
	}, nil
}

// ValidateForCreate 校验创建入参；status 未传时默认草稿。
func (r ArticleWriteRequest) ValidateForCreate() (Article, error) {
	article, err := r.normalize()
	if err != nil {
		return Article{}, err
	}
	status, err := resolveArticleStatus(r.Status, ArticleStatusDraft)
	if err != nil {
		return Article{}, err
	}
	article.Status = status
	return article, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r ArticleWriteRequest) ValidateForUpdate(existing Article) (Article, error) {
	article, err := r.normalize()
	if err != nil {
		return Article{}, err
	}
	status, err := resolveArticleStatus(r.Status, existing.Status)
	if err != nil {
		return Article{}, err
	}
	article.Status = status
	return article, nil
}

// UpdateMap 返回 GORM Updates 用的字段 map（支持零值写入；不含 publishedAt 与计数）。
func (a Article) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"title":            a.Title,
		"slug":             a.Slug,
		"textbook_version": a.TextbookVersion,
		"grade":            a.Grade,
		"semester":         a.Semester,
		"unit":             a.Unit,
		"essay_topic":      a.EssayTopic,
		"content_type":     a.ContentType,
		"audience":         a.Audience,
		"summary":          a.Summary,
		"cover_img":        a.CoverImg,
		"body_images":      a.BodyImages,
		"status":           a.Status,
		"is_recommended":   a.IsRecommended,
		"is_hot":           a.IsHot,
		"sort_order":       a.SortOrder,
	}
}

// CheckArticleSlugDuplicate 检查 slug 是否与其他文章冲突（slug 为空时跳过）。
func CheckArticleSlugDuplicate(db *gorm.DB, slug string, excludeID uint) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil
	}
	q := db.Model(&Article{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("article slug already exists")
	}
	return nil
}
