// 标签 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package admin

import (
	"fmt"
	"strings"

	"course_server/internal/dto"
	"course_server/internal/model"
)

// TagWriteRequest 标签创建/更新共用入参；仅需 name，type 已废弃。
type TagWriteRequest struct {
	Name      string `json:"name" binding:"required"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

// CreateTagRequest 创建标签请求。
type CreateTagRequest = TagWriteRequest

// UpdateTagRequest 更新标签请求。
type UpdateTagRequest = TagWriteRequest

const maxTagNameLen = 80

func (r TagWriteRequest) normalize() (string, error) {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	if len(name) > maxTagNameLen {
		return "", fmt.Errorf("name exceeds max %d", maxTagNameLen)
	}
	return name, nil
}

// ValidateForCreate 校验创建入参；status 未传时默认启用。
func (r TagWriteRequest) ValidateForCreate() (model.Tag, error) {
	name, err := r.normalize()
	if err != nil {
		return model.Tag{}, err
	}
	status, err := dto.ResolveStatus(r.Status, model.StatusEnabled)
	if err != nil {
		return model.Tag{}, err
	}
	return model.Tag{Name: name, SortOrder: r.SortOrder, Status: status}, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r TagWriteRequest) ValidateForUpdate(existing model.Tag) (model.Tag, error) {
	name, err := r.normalize()
	if err != nil {
		return model.Tag{}, err
	}
	status, err := dto.ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return model.Tag{}, err
	}
	return model.Tag{Name: name, SortOrder: r.SortOrder, Status: status}, nil
}
