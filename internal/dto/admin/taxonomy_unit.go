// 教材分类单元 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package admin

import (
	"fmt"
	"strings"

	"course_server/internal/dto"
	"course_server/internal/model"
)

// TaxonomyUnitWriteRequest 分类单元创建/更新共用入参字段。
type TaxonomyUnitWriteRequest struct {
	TextbookVersion string `json:"textbookVersion"`
	Grade           int    `json:"grade" binding:"required"`
	Semester        string `json:"semester" binding:"required"`
	Unit            int    `json:"unit" binding:"required"`
	UnitTitle       string `json:"unitTitle"`
	SortOrder       *int   `json:"sortOrder"`
	Status          *int   `json:"status"`
}

// CreateTaxonomyUnitRequest 创建分类单元请求。
type CreateTaxonomyUnitRequest = TaxonomyUnitWriteRequest

// UpdateTaxonomyUnitRequest 更新分类单元请求。
type UpdateTaxonomyUnitRequest = TaxonomyUnitWriteRequest

// normalize 校验入参字段；textbookVersion 默认值由 ValidateForCreate/ValidateForUpdate 分别处理。
func (r TaxonomyUnitWriteRequest) normalize() (model.TaxonomyUnit, error) {
	version := strings.TrimSpace(r.TextbookVersion)
	semester := strings.TrimSpace(r.Semester)
	if semester == "" {
		return model.TaxonomyUnit{}, fmt.Errorf("semester is required")
	}
	if !model.ValidSemester(semester) {
		return model.TaxonomyUnit{}, fmt.Errorf("invalid semester")
	}
	if r.Grade <= 0 {
		return model.TaxonomyUnit{}, fmt.Errorf("grade must be greater than 0")
	}
	if r.Unit <= 0 {
		return model.TaxonomyUnit{}, fmt.Errorf("unit must be greater than 0")
	}
	sortOrder := 0
	if r.SortOrder != nil {
		sortOrder = *r.SortOrder
	}
	return model.TaxonomyUnit{
		TextbookVersion: version,
		Grade:           r.Grade,
		Semester:        semester,
		Unit:            r.Unit,
		UnitTitle:       strings.TrimSpace(r.UnitTitle),
		SortOrder:       sortOrder,
	}, nil
}

// ValidateForCreate 校验创建入参；status 未传时默认启用。
func (r TaxonomyUnitWriteRequest) ValidateForCreate() (model.TaxonomyUnit, error) {
	unit, err := r.normalize()
	if err != nil {
		return model.TaxonomyUnit{}, err
	}
	status, err := dto.ResolveStatus(r.Status, model.StatusEnabled)
	if err != nil {
		return model.TaxonomyUnit{}, err
	}
	unit.Status = status
	if unit.TextbookVersion == "" {
		unit.TextbookVersion = model.DefaultTextbookVersion
	}
	return unit, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r TaxonomyUnitWriteRequest) ValidateForUpdate(existing model.TaxonomyUnit) (model.TaxonomyUnit, error) {
	unit, err := r.normalize()
	if err != nil {
		return model.TaxonomyUnit{}, err
	}
	status, err := dto.ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return model.TaxonomyUnit{}, err
	}
	unit.Status = status
	if unit.TextbookVersion == "" {
		unit.TextbookVersion = strings.TrimSpace(existing.TextbookVersion)
		if unit.TextbookVersion == "" {
			unit.TextbookVersion = model.DefaultTextbookVersion
		}
	}
	if r.SortOrder == nil {
		unit.SortOrder = existing.SortOrder
	}
	return unit, nil
}
