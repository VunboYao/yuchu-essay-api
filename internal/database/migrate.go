package database

import (
	"errors"

	"course_server/internal/model"
)

// AutoMigrate 根据 model 定义创建/更新全部表结构（Day2）。
func AutoMigrate() error {
	if DB == nil {
		return errors.New("database is not initialized")
	}
	return DB.AutoMigrate(
		&model.Admin{},
		&model.AdminOperationLog{},
		&model.User{},
		&model.TaxonomyUnit{},
		&model.Tag{},
		&model.Article{},
		&model.ArticleTag{},
		&model.EssayExample{},
		&model.EssayExampleTag{},
		&model.DownloadFile{},
		&model.DownloadLog{},
		&model.Favorite{},
		&model.ViewLog{},
		&model.SearchLog{},
		&model.ShareLog{},
		&model.AdSlot{},
		&model.AdEvent{},
		&model.HomeBlock{},
		&model.OfficialAccountConfig{},
		&model.SystemSetting{},
	)
}
