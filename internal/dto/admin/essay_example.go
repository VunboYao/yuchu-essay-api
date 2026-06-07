// 优秀作文 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package admin

import (
	"fmt"
	"strings"

	"course_server/internal/dto"
	"course_server/internal/model"
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

func (r EssayExampleWriteRequest) normalize() (model.EssayExample, error) {
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return model.EssayExample{}, fmt.Errorf("title is required")
	}
	return model.EssayExample{
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
func (r EssayExampleWriteRequest) ValidateForCreate() (model.EssayExample, error) {
	essay, err := r.normalize()
	if err != nil {
		return model.EssayExample{}, err
	}
	status, err := dto.ResolveStatus(r.Status, model.StatusEnabled)
	if err != nil {
		return model.EssayExample{}, err
	}
	essay.Status = status
	return essay, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r EssayExampleWriteRequest) ValidateForUpdate(existing model.EssayExample) (model.EssayExample, error) {
	essay, err := r.normalize()
	if err != nil {
		return model.EssayExample{}, err
	}
	status, err := dto.ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return model.EssayExample{}, err
	}
	essay.Status = status
	return essay, nil
}
