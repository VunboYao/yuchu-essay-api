// Admin 写操作通用辅助：Updates(map) + 重新查库，供各 CRUD handler 复用。
package handler

import "gorm.io/gorm"

// UpdateAndReload 对 dest 指向的记录执行 Updates，并重新查库刷新 dest。
// dest 须已加载且含有效主键（通常为先 First 查出的 &item）。
func UpdateAndReload(db *gorm.DB, dest interface{}, updates map[string]interface{}) error {
	if err := db.Model(dest).Updates(updates).Error; err != nil {
		return err
	}
	return db.First(dest).Error
}
