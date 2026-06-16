// 我的收藏列表：读时组装文章/范文展示摘要（favorites 表仍只存引用）。
package handler

import (
	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/jsontime"

	"gorm.io/gorm"
)

// FavoriteListItem 收藏列表项：关系字段 + 是否可访问 + 目标摘要。
type FavoriteListItem struct {
	ID         uint          `json:"id"`
	TargetType string        `json:"targetType"`
	TargetID   uint          `json:"targetId"`
	CreatedAt  jsontime.Time `json:"createdAt"`
	Available  bool          `json:"available"`
	Target     any           `json:"target"`
}

// loadFavoriteListPage 分页查收藏并批量组装展示数据。
func loadFavoriteListPage(userID uint, page, pageSize int) ([]FavoriteListItem, int64, error) {
	db := database.DB.Model(&model.Favorite{}).Where("user_id = ?", userID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var favs []model.Favorite
	if err := paginate(db.Order("id DESC"), page, pageSize).Find(&favs).Error; err != nil {
		return nil, 0, err
	}
	if len(favs) == 0 {
		return []FavoriteListItem{}, total, nil
	}
	articles, essays, linkedArticles, err := loadFavoriteTargets(database.DB, favs)
	if err != nil {
		return nil, 0, err
	}
	return BuildFavoriteListItems(favs, articles, essays, linkedArticles), total, nil
}

func loadFavoriteTargets(db *gorm.DB, favs []model.Favorite) (
	map[uint]model.Article,
	map[uint]model.EssayExample,
	map[uint]ArticleTitleRef,
	error,
) {
	refs := make([]contentTargetRef, len(favs))
	for i, fav := range favs {
		refs[i] = contentTargetRef{TargetType: fav.TargetType, TargetID: fav.TargetID}
	}
	articleIDs, essayIDs := collectContentTargetIDs(refs)
	return loadContentTargets(db, articleIDs, essayIDs)
}

// BuildFavoriteListItems 将收藏记录与批量查到的目标合并为列表项（保持收藏顺序）。
func BuildFavoriteListItems(
	favs []model.Favorite,
	articles map[uint]model.Article,
	essays map[uint]model.EssayExample,
	linkedArticles map[uint]ArticleTitleRef,
) []FavoriteListItem {
	items := make([]FavoriteListItem, 0, len(favs))
	for _, fav := range favs {
		available, target := buildContentTarget(fav.TargetType, fav.TargetID, articles, essays, linkedArticles)
		items = append(items, FavoriteListItem{
			ID:         fav.ID,
			TargetType: fav.TargetType,
			TargetID:   fav.TargetID,
			CreatedAt:  fav.CreatedAt,
			Available:  available,
			Target:     target,
		})
	}
	return items
}
