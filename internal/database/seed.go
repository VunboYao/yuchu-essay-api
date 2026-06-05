package database

import (
	"errors"

	"course_server/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// SeedAdmin 初始化默认管理员与基础标签（Day2/Day3）。
// 管理员：admin/admin123；标签含「免费」供下载权限校验使用。
// 注意：taxonomy_units 未在此种子，需后台或后续脚本维护。
func SeedAdmin() error {
	if DB == nil {
		return errors.New("database is not initialized")
	}

	var count int64
	// 检查是否已经存在管理员
	if err := DB.Model(&model.Admin{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return err
	}
	// 如果管理员不存在，则创建管理员
	if count == 0 {
		// 生成密码哈希
		hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		// 创建管理员
		if err := DB.Create(&model.Admin{
			Username:     "admin",
			PasswordHash: string(hash),
			Role:         "admin",
			Status:       model.StatusEnabled,
		}).Error; err != nil {
			return err
		}
	}

	// 创建基础标签
	seedTags := []model.Tag{
		{Name: "免费", Type: "download", Status: model.StatusEnabled},
		{Name: "想象作文", Type: "topic", Status: model.StatusEnabled},
		{Name: "优秀范文", Type: "essay", Status: model.StatusEnabled},
	}
	// 创建基础标签
	for _, tag := range seedTags {
		DB.FirstOrCreate(&tag, model.Tag{Name: tag.Name, Type: tag.Type})
	}

	return nil
}
