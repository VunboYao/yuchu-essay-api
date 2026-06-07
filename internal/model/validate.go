package model

import (
	"gorm.io/gorm"
)

// NextSortOrder 查询指定模型表 sort_order 最大值并返回 +1，用于创建时未传 sortOrder 自动递增。
func NextSortOrder(db *gorm.DB, m interface{}) int {
	var maxOrder int
	db.Model(m).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)
	return maxOrder + 1
}
