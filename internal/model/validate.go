package model

import (
	"fmt"

	"gorm.io/gorm"
)

// ResolveStatus 解析可选 status：nil 时用 defaultStatus；非 nil 时须为 StatusDisabled 或 StatusEnabled。
func ResolveStatus(ptr *int, defaultStatus int) (int, error) {
	if ptr == nil {
		return defaultStatus, nil
	}
	if *ptr != StatusDisabled && *ptr != StatusEnabled {
		return 0, fmt.Errorf("invalid status")
	}
	return *ptr, nil
}

// NextSortOrder 查询指定模型表 sort_order 最大值并返回 +1，用于创建时未传 sortOrder 自动递增。
func NextSortOrder(db *gorm.DB, m interface{}) int {
	var maxOrder int
	db.Model(m).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)
	return maxOrder + 1
}
