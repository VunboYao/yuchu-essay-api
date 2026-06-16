// 浏览历史列表：读时组装文章/范文展示摘要（view_logs 表仍只存引用）。
package handler

import (
	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/jsontime"

	"gorm.io/gorm"
)

// HistoryListItem 浏览历史列表项。
type HistoryListItem struct {
	ID          uint          `json:"id"`
	TargetType  string        `json:"targetType"`
	TargetID    uint          `json:"targetId"`
	PagePath    string        `json:"pagePath"`
	SourceScene string        `json:"sourceScene"`
	CreatedAt   jsontime.Time `json:"createdAt"`
	Available   bool          `json:"available"`
	Target      any           `json:"target"`
}

func loadHistoryListPage(userID uint, page, pageSize int) ([]HistoryListItem, int64, error) {
	db := database.DB.Model(&model.ViewLog{}).Where("user_id = ?", userID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []model.ViewLog
	if err := paginate(db.Order("created_at DESC"), page, pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	if len(logs) == 0 {
		return []HistoryListItem{}, total, nil
	}
	articles, essays, linkedArticles, err := loadHistoryTargets(database.DB, logs)
	if err != nil {
		return nil, 0, err
	}
	return BuildHistoryListItems(logs, articles, essays, linkedArticles), total, nil
}

func loadHistoryTargets(db *gorm.DB, logs []model.ViewLog) (
	map[uint]model.Article,
	map[uint]model.EssayExample,
	map[uint]ArticleTitleRef,
	error,
) {
	refs := make([]contentTargetRef, len(logs))
	for i, log := range logs {
		refs[i] = contentTargetRef{TargetType: log.TargetType, TargetID: log.TargetID}
	}
	articleIDs, essayIDs := collectContentTargetIDs(refs)
	return loadContentTargets(db, articleIDs, essayIDs)
}

// BuildHistoryListItems 将浏览记录与批量查到的目标合并为列表项。
func BuildHistoryListItems(
	logs []model.ViewLog,
	articles map[uint]model.Article,
	essays map[uint]model.EssayExample,
	linkedArticles map[uint]ArticleTitleRef,
) []HistoryListItem {
	items := make([]HistoryListItem, 0, len(logs))
	for _, log := range logs {
		available, target := buildContentTarget(log.TargetType, log.TargetID, articles, essays, linkedArticles)
		items = append(items, HistoryListItem{
			ID:          log.ID,
			TargetType:  log.TargetType,
			TargetID:    log.TargetID,
			PagePath:    log.PagePath,
			SourceScene: log.SourceScene,
			CreatedAt:   log.CreatedAt,
			Available:   available,
			Target:      target,
		})
	}
	return items
}
