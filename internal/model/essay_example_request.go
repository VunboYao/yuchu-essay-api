// 优秀作文 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// EssayExampleWriteRequest 优秀作文创建/更新共用入参；不含 id、时间戳与计数。
type EssayExampleWriteRequest struct {
	ArticleID uint   `json:"articleId"`
	Title     string `json:"title" binding:"required"`
	Grade     int    `json:"grade"`
	WordCount int    `json:"wordCount"`
	Body      string `json:"body"`
	Comment   string `json:"comment"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

// CreateEssayExampleRequest 创建优秀作文请求。
type CreateEssayExampleRequest = EssayExampleWriteRequest

// UpdateEssayExampleRequest 更新优秀作文请求。
type UpdateEssayExampleRequest = EssayExampleWriteRequest

func (r EssayExampleWriteRequest) normalize() (EssayExample, error) {
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return EssayExample{}, fmt.Errorf("title is required")
	}
	return EssayExample{
		ArticleID: r.ArticleID,
		Title:     title,
		Grade:     r.Grade,
		WordCount: r.WordCount,
		Body:      r.Body,
		Comment:   strings.TrimSpace(r.Comment),
		SortOrder: r.SortOrder,
	}, nil
}

// ValidateForCreate 校验创建入参；status 未传时默认启用。
func (r EssayExampleWriteRequest) ValidateForCreate() (EssayExample, error) {
	essay, err := r.normalize()
	if err != nil {
		return EssayExample{}, err
	}
	status, err := ResolveStatus(r.Status, StatusEnabled)
	if err != nil {
		return EssayExample{}, err
	}
	essay.Status = status
	return essay, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r EssayExampleWriteRequest) ValidateForUpdate(existing EssayExample) (EssayExample, error) {
	essay, err := r.normalize()
	if err != nil {
		return EssayExample{}, err
	}
	status, err := ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return EssayExample{}, err
	}
	essay.Status = status
	return essay, nil
}

// UpdateMap 返回 GORM Updates 用的字段 map（支持零值写入；不含计数）。
func (e EssayExample) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"article_id": e.ArticleID,
		"title":      e.Title,
		"grade":      e.Grade,
		"word_count": e.WordCount,
		"body":       e.Body,
		"comment":    e.Comment,
		"sort_order": e.SortOrder,
		"status":     e.Status,
	}
}

// CheckEssayArticleExists 校验关联文章是否存在（articleId 为 0 时跳过）。
func CheckEssayArticleExists(db *gorm.DB, articleID uint) error {
	if articleID == 0 {
		return nil
	}
	var count int64
	if err := db.Model(&Article{}).Where("id = ?", articleID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("article not found")
	}
	return nil
}
