// 教材分类单元 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// 分类单元创建/更新共用入参字段。
type TaxonomyUnitWriteRequest struct {
	TextbookVersion string `json:"textbookVersion"`
	Grade           int    `json:"grade" binding:"required"`
	Semester        string `json:"semester" binding:"required"`
	Unit            int    `json:"unit" binding:"required"`
	UnitTitle       string `json:"unitTitle"`
	SortOrder       *int   `json:"sortOrder"`
	Status          *int   `json:"status"`
}

// 创建分类单元请求。
type CreateTaxonomyUnitRequest = TaxonomyUnitWriteRequest

// 更新分类单元请求。
type UpdateTaxonomyUnitRequest = TaxonomyUnitWriteRequest

// normalize 校验并补全入参字段。
func (r TaxonomyUnitWriteRequest) normalize() (TaxonomyUnit, error) {
	// 版本号字段校验
	version := strings.TrimSpace(r.TextbookVersion)
	if version == "" {
		version = DefaultTextbookVersion
	}
	// 学期须为上册/下册之一
	semester := strings.TrimSpace(r.Semester)
	if semester == "" {
		return TaxonomyUnit{}, fmt.Errorf("semester is required")
	}
	if !ValidSemester(semester) {
		return TaxonomyUnit{}, fmt.Errorf("invalid semester")
	}
	// 年级
	if r.Grade <= 0 {
		return TaxonomyUnit{}, fmt.Errorf("grade must be greater than 0")
	}
	// 单元
	if r.Unit <= 0 {
		return TaxonomyUnit{}, fmt.Errorf("unit must be greater than 0")
	}
	sortOrder := 0
	if r.SortOrder != nil {
		sortOrder = *r.SortOrder
	}
	return TaxonomyUnit{
		TextbookVersion: version,
		Grade:           r.Grade,
		Semester:        semester,
		Unit:            r.Unit,
		UnitTitle:       strings.TrimSpace(r.UnitTitle),
		SortOrder:       sortOrder,
	}, nil
}

// ValidateForCreate 校验创建入参；status 未传时默认启用。
func (r TaxonomyUnitWriteRequest) ValidateForCreate() (TaxonomyUnit, error) {
	unit, err := r.normalize()
	if err != nil {
		return TaxonomyUnit{}, err
	}
	// 如果请求未传 status 字段，默认设为启用状态（StatusEnabled）
	status, err := ResolveStatus(r.Status, StatusEnabled)
	if err != nil {
		return TaxonomyUnit{}, err
	}
	unit.Status = status
	return unit, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r TaxonomyUnitWriteRequest) ValidateForUpdate(existing TaxonomyUnit) (TaxonomyUnit, error) {
	unit, err := r.normalize()
	if err != nil {
		return TaxonomyUnit{}, err
	}
	status, err := ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return TaxonomyUnit{}, err
	}
	unit.Status = status
	// sortOrder 未传时保留原值
	if r.SortOrder == nil {
		unit.SortOrder = existing.SortOrder
	}
	return unit, nil
}

// UpdateMap 返回 GORM Updates 用的字段 map（支持零值写入）。
func (t TaxonomyUnit) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"textbook_version": t.TextbookVersion,
		"grade":            t.Grade,
		"semester":         t.Semester,
		"unit":             t.Unit,
		"unit_title":       t.UnitTitle,
		"sort_order":       t.SortOrder,
		"status":           t.Status,
	}
}

// CheckTaxonomyUnitDuplicate 检查同版本/年级/学期/单元组合是否已存在。
func CheckTaxonomyUnitDuplicate(db *gorm.DB, unit TaxonomyUnit, excludeID uint) error {
	q := db.Model(&TaxonomyUnit{}).Where(
		"textbook_version = ? AND grade = ? AND semester = ? AND unit = ?",
		unit.TextbookVersion, unit.Grade, unit.Semester, unit.Unit,
	)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("taxonomy unit with same textbook version, grade, semester and unit already exists")
	}
	return nil
}
