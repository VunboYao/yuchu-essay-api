// 文章/范文列表展示摘要：收藏、浏览历史等读接口共用。
package handler

import (
	"course_server/internal/model"

	"gorm.io/gorm"
)

// ContentArticleSummary 文章列表卡片字段（不含正文图）。
type ContentArticleSummary struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	CoverImg   string `json:"coverImg"`
	Grade      int    `json:"grade"`
	Semester   string `json:"semester"`
	Unit       int    `json:"unit"`
	EssayTopic string `json:"essayTopic"`
	Summary    string `json:"summary"`
}

// ContentEssaySummary 范文列表卡片字段（无 coverImg，前端用默认图）。
type ContentEssaySummary struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	Grade        int    `json:"grade"`
	WordCount    int    `json:"wordCount"`
	ArticleID    *uint  `json:"articleId,omitempty"`
	ArticleTitle string `json:"articleTitle,omitempty"`
}

// FavoriteArticleSummary 兼容收藏列表类型名。
type FavoriteArticleSummary = ContentArticleSummary

// FavoriteEssaySummary 兼容收藏列表类型名。
type FavoriteEssaySummary = ContentEssaySummary

// ArticleTitleRef 专题 id/title，供范文展示所属专题。
type ArticleTitleRef struct {
	ID    uint   `gorm:"column:id"`
	Title string `gorm:"column:title"`
}

type contentTargetRef struct {
	TargetType string
	TargetID   uint
}

// loadContentTargets 批量加载文章、范文及范文关联专题标题。
func loadContentTargets(db *gorm.DB, articleIDs, essayIDs []uint) (
	map[uint]model.Article,
	map[uint]model.EssayExample,
	map[uint]ArticleTitleRef,
	error,
) {
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

func collectContentTargetIDs(refs []contentTargetRef) (articleIDs, essayIDs []uint) {
	seenArticle := map[uint]bool{}
	seenEssay := map[uint]bool{}
	for _, ref := range refs {
		switch ref.TargetType {
		case model.TargetTypeArticle:
			if !seenArticle[ref.TargetID] {
				seenArticle[ref.TargetID] = true
				articleIDs = append(articleIDs, ref.TargetID)
			}
		case model.TargetTypeEssay:
			if !seenEssay[ref.TargetID] {
				seenEssay[ref.TargetID] = true
				essayIDs = append(essayIDs, ref.TargetID)
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

// buildContentTarget 按类型组装可访问性与展示摘要。
func buildContentTarget(
	targetType string,
	targetID uint,
	articles map[uint]model.Article,
	essays map[uint]model.EssayExample,
	linkedArticles map[uint]ArticleTitleRef,
) (available bool, target any) {
	switch targetType {
	case model.TargetTypeArticle:
		article, ok := articles[targetID]
		if !ok || article.Status != model.ArticleStatusPublished {
			return false, nil
		}
		return true, ContentArticleSummary{
			ID:         article.ID,
			Title:      article.Title,
			CoverImg:   article.CoverImg,
			Grade:      article.Grade,
			Semester:   article.Semester,
			Unit:       article.Unit,
			EssayTopic: article.EssayTopic,
			Summary:    article.Summary,
		}
	case model.TargetTypeEssay:
		essay, ok := essays[targetID]
		if !ok || essay.Status != model.StatusEnabled {
			return false, nil
		}
		summary := ContentEssaySummary{
			ID:        essay.ID,
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
		return true, summary
	default:
		return false, nil
	}
}
