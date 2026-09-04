package database

import (
	"errors"

	"course_server/internal/model"
)

// AutoMigrate 根据 model 创建/更新表结构，并写入表级 COMMENT（Day2）。
// 迁移顺序与 internal/model 分文件一致：共用 → 后台 → 小程序；二期订单见 model_order.go。
func AutoMigrate() error {
	if DB == nil {
		return errors.New("database is not initialized")
	}
	if err := DB.AutoMigrate(
		// 【共用】内容/配置
		&model.TaxonomyUnit{},
		&model.Tag{},
		&model.Article{},
		&model.ArticleTag{},
		&model.EssayExample{},
		&model.EssayExampleTag{},
		&model.Video{},
		&model.VideoTag{},
		&model.DownloadFile{},
		&model.AdSlot{},
		&model.HomeBlock{},
		&model.HotSearchKeyword{},
		&model.OfficialAccountConfig{},
		// 【后台】
		&model.Admin{},
		&model.AdminOperationLog{},
		&model.SystemSetting{},
		// 【小程序】
		&model.User{},
		&model.DownloadLog{},
		&model.Favorite{},
		&model.ViewLog{},
		&model.SearchLog{},
		&model.ShareLog{},
		&model.AdEvent{},
		&model.Feedback{},
	); err != nil {
		return err
	}
	return model.ApplyTableComments(DB)
}
