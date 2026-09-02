// Admin 写操作通用辅助：Create/Updates 走 map 以写入零值，供各 CRUD handler 复用。
package handler

import (
	"fmt"

	"gorm.io/gorm"
)

// UpdateAndReload 对 dest 指向的记录执行 Updates，并重新查库刷新 dest。
// dest 须已加载且含有效主键（通常为先 First 查出的 &item）。
func UpdateAndReload(db *gorm.DB, dest interface{}, updates map[string]interface{}) error {
	if err := db.Model(dest).Updates(updates).Error; err != nil {
		return err
	}
	return db.First(dest).Error
}

// CreateAndReload 按 map 插入 dest 并重新查库。
// 普通 Create(struct) 会跳过带 gorm default 的零值（status=0 会落成列 DEFAULT 1）；map 与 UpdateMap 同一口径。
func CreateAndReload(db *gorm.DB, dest interface{}, values map[string]interface{}) error {
	if err := db.Model(dest).Create(values).Error; err != nil {
		return err
	}
	id, ok := values["id"]
	if !ok {
		return fmt.Errorf("create: missing id")
	}
	return db.First(dest, id).Error
}
