// 优秀作文 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package admin

import (
	"fmt"
	"strings"
	"unicode"

	"course_server/internal/dto"
	"course_server/internal/model"
)

// EssayExampleWriteRequest 优秀作文创建/更新共用入参；不含 id、时间戳与计数。
type EssayExampleWriteRequest struct {
	ArticleID *uint  `json:"articleId"`
	Title     string `json:"title" binding:"required"`
	Grade     int    `json:"grade" binding:"required"`
	Body      string `json:"body" binding:"required"`
	Comment   string `json:"comment" binding:"required"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
	TagIDs    []uint `json:"tagIds"`
}

// CreateEssayExampleRequest 创建优秀作文请求。
type CreateEssayExampleRequest = EssayExampleWriteRequest

// UpdateEssayExampleRequest 更新优秀作文请求。
type UpdateEssayExampleRequest = EssayExampleWriteRequest

// resolveOptionalArticleID 解析可选专题关联：未传、null、0 均视为不关联。
func resolveOptionalArticleID(ptr *uint) *uint {
	if ptr == nil || *ptr == 0 {
		return nil
	}
	return ptr
}

// countEssayWordCount 按正文字符统计字数（不计空白符，与中文「字数」习惯一致）。
func countEssayWordCount(body string) int {
	n := 0
	for _, r := range body {
		if !unicode.IsSpace(r) {
			n++
		}
	}
	return n
}

func (r EssayExampleWriteRequest) normalize() (model.EssayExample, error) {
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return model.EssayExample{}, fmt.Errorf("title is required")
	}
	if r.Grade <= 0 {
		return model.EssayExample{}, fmt.Errorf("grade must be greater than 0")
	}
	body := strings.TrimSpace(r.Body)
	if body == "" {
		return model.EssayExample{}, fmt.Errorf("body is required")
	}
	comment := strings.TrimSpace(r.Comment)
	if comment == "" {
		return model.EssayExample{}, fmt.Errorf("comment is required")
	}
	return model.EssayExample{
		ArticleID: resolveOptionalArticleID(r.ArticleID),
		Title:     title,
		Grade:     r.Grade,
		WordCount: countEssayWordCount(body),
		Body:      body,
		Comment:   comment,
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
