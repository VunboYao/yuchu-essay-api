// 标签 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// TagWriteRequest 标签创建/更新共用入参字段。
type TagWriteRequest struct {
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

// CreateTagRequest 创建标签请求。
type CreateTagRequest = TagWriteRequest

// UpdateTagRequest 更新标签请求。
type UpdateTagRequest = TagWriteRequest

func (r TagWriteRequest) normalize() (name, typ string, err error) {
	name = strings.TrimSpace(r.Name)
	typ = strings.TrimSpace(r.Type)
	if name == "" {
		return "", "", fmt.Errorf("name is required")
	}
	if typ == "" {
		return "", "", fmt.Errorf("type is required")
	}
	if !ValidTagType(typ) {
		return "", "", fmt.Errorf("invalid tag type")
	}
	return name, typ, nil
}

// ValidateForCreate 校验创建入参；status 未传时默认启用。
func (r TagWriteRequest) ValidateForCreate() (Tag, error) {
	name, typ, err := r.normalize()
	if err != nil {
		return Tag{}, err
	}
	status, err := ResolveStatus(r.Status, StatusEnabled)
	if err != nil {
		return Tag{}, err
	}
	return Tag{Name: name, Type: typ, SortOrder: r.SortOrder, Status: status}, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r TagWriteRequest) ValidateForUpdate(existing Tag) (Tag, error) {
	name, typ, err := r.normalize()
	if err != nil {
		return Tag{}, err
	}
	status, err := ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return Tag{}, err
	}
	return Tag{Name: name, Type: typ, SortOrder: r.SortOrder, Status: status}, nil
}

// UpdateMap 返回 GORM Updates 用的字段 map（支持零值写入）。
func (t Tag) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"name":       t.Name,
		"type":       t.Type,
		"sort_order": t.SortOrder,
		"status":     t.Status,
	}
}

// CheckTagNameTypeDuplicate 检查 name+type 是否与其他记录冲突。
func CheckTagNameTypeDuplicate(db *gorm.DB, name, typ string, excludeID uint) error {
	q := db.Model(&Tag{}).Where("name = ? AND type = ?", name, typ)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("tag with same name and type already exists")
	}
	return nil
}
