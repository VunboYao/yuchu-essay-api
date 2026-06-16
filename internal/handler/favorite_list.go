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

// FavoriteArticleSummary 文章收藏卡片字段（不含正文图）。
type FavoriteArticleSummary struct {
	Title      string `json:"title"`
	CoverImg   string `json:"coverImg"`
	Grade      int    `json:"grade"`
	Semester   string `json:"semester"`
	Unit       int    `json:"unit"`
	EssayTopic string `json:"essayTopic"`
	Summary    string `json:"summary"`
}

// FavoriteEssaySummary 范文收藏卡片字段（无 coverImg，前端用默认图）。
type FavoriteEssaySummary struct {
	Title        string `json:"title"`
	Grade        int    `json:"grade"`
	WordCount    int    `json:"wordCount"`
	ArticleID    *uint  `json:"articleId,omitempty"`
	ArticleTitle string `json:"articleTitle,omitempty"`
}

// ArticleTitleRef 专题 id/title，供收藏列表展示范文所属专题。
type ArticleTitleRef struct {
	ID    uint   `gorm:"column:id"`
	Title string `gorm:"column:title"`
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

// loadFavoriteTargets 按收藏页批量加载文章、范文及范文关联专题标题。
func loadFavoriteTargets(db *gorm.DB, favs []model.Favorite) (
	map[uint]model.Article,
	map[uint]model.EssayExample,
	map[uint]ArticleTitleRef,
	error,
) {
	articleIDs, essayIDs := favoriteTargetIDs(favs)
	articles := map[uint]model.Article{}
	essays := map[uint]model.EssayExample{}
	linkedArticles := map[uint]ArticleTitleRef{}

	if len(articleIDs) > 0 {
		var rows []model.Article
		if err := db.Select("id", "title", "cover_img", "grade", "semester", "unit", "essay_topic", "summary", "status").
			Where("id IN ?", articleIDs).Find(&rows).Error; err != nil {
			return nil, nil, nil, err
		}
		for _, row := range rows {
			articles[row.ID] = row
		}
	}
	if len(essayIDs) > 0 {
		var rows []model.EssayExample
		if err := db.Select("id", "title", "grade", "word_count", "article_id", "status").
			Where("id IN ?", essayIDs).Find(&rows).Error; err != nil {
			return nil, nil, nil, err
		}
		for _, row := range rows {
			essays[row.ID] = row
		}
	}
	linkedArticleIDs := linkedArticleIDsFromEssays(essays)
	if len(linkedArticleIDs) > 0 {
		var rows []ArticleTitleRef
		if err := db.Model(&model.Article{}).Select("id", "title").
			Where("id IN ?", linkedArticleIDs).Find(&rows).Error; err != nil {
			return nil, nil, nil, err
		}
		for _, row := range rows {
			linkedArticles[row.ID] = row
		}
	}
	return articles, essays, linkedArticles, nil
}

func favoriteTargetIDs(favs []model.Favorite) (articleIDs, essayIDs []uint) {
	seenArticle := map[uint]bool{}
	seenEssay := map[uint]bool{}
	for _, fav := range favs {
		switch fav.TargetType {
		case model.TargetTypeArticle:
			if !seenArticle[fav.TargetID] {
				seenArticle[fav.TargetID] = true
				articleIDs = append(articleIDs, fav.TargetID)
			}
		case model.TargetTypeEssay:
			if !seenEssay[fav.TargetID] {
				seenEssay[fav.TargetID] = true
				essayIDs = append(essayIDs, fav.TargetID)
			}
		}
	}
	return articleIDs, essayIDs
}

func linkedArticleIDsFromEssays(essays map[uint]model.EssayExample) []uint {
	seen := map[uint]bool{}
	var ids []uint
	for _, essay := range essays {
		if essay.ArticleID == nil || *essay.ArticleID == 0 || seen[*essay.ArticleID] {
			continue
		}
		seen[*essay.ArticleID] = true
		ids = append(ids, *essay.ArticleID)
	}
	return ids
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
		item := FavoriteListItem{
			ID:         fav.ID,
			TargetType: fav.TargetType,
			TargetID:   fav.TargetID,
			CreatedAt:  fav.CreatedAt,
		}
		switch fav.TargetType {
		case model.TargetTypeArticle:
			article, ok := articles[fav.TargetID]
			if !ok || article.Status != model.ArticleStatusPublished {
				item.Available = false
				item.Target = nil
			} else {
				item.Available = true
				item.Target = FavoriteArticleSummary{
					Title:      article.Title,
					CoverImg:   article.CoverImg,
					Grade:      article.Grade,
					Semester:   article.Semester,
					Unit:       article.Unit,
					EssayTopic: article.EssayTopic,
					Summary:    article.Summary,
				}
			}
		case model.TargetTypeEssay:
			essay, ok := essays[fav.TargetID]
			if !ok || essay.Status != model.StatusEnabled {
				item.Available = false
				item.Target = nil
			} else {
				item.Available = true
				summary := FavoriteEssaySummary{
					Title:     essay.Title,
					Grade:     essay.Grade,
					WordCount: essay.WordCount,
				}
				if essay.ArticleID != nil && *essay.ArticleID > 0 {
					summary.ArticleID = essay.ArticleID
					if ref, ok := linkedArticles[*essay.ArticleID]; ok {
						summary.ArticleTitle = ref.Title
					}
				}
				item.Target = summary
			}
		default:
			item.Available = false
			item.Target = nil
		}
		items = append(items, item)
	}
	return items
}
