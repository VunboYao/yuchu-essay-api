// 后台内容标签关联内部 helper（article_tags / essay_example_tags）。
// 标签读写已并入文章/优秀作文 CRUD 的 tagIds 字段，采用先清后插的全量替换语义。
package handler

import (
	"course_server/internal/database"
	"course_server/internal/model"

	"gorm.io/gorm"
)

// dedupeTagIDs 去重 tag id，保持首次出现顺序。
func dedupeTagIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return ids
	}
	seen := map[uint]bool{}
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// validTagIDs 过滤出实际存在的标签 ID，避免写入无效关联（本表无外键约束）。
func validTagIDs(db *gorm.DB, ids []uint) []uint {
	if len(ids) == 0 {
		return nil
	}
	var existing []uint
	db.Model(&model.Tag{}).Where("id IN ?", ids).Pluck("id", &existing)
	return existing
}

// replaceArticleTags 全量替换文章标签：先删后批量插入。
func replaceArticleTags(tx *gorm.DB, articleID uint, tagIDs []uint) error {
	tagIDs = validTagIDs(tx, dedupeTagIDs(tagIDs))
	if err := tx.Where("article_id = ?", articleID).Delete(&model.ArticleTag{}).Error; err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	rows := make([]model.ArticleTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		rows = append(rows, model.ArticleTag{ArticleID: articleID, TagID: tagID})
	}
	return tx.Create(&rows).Error
}

// replaceEssayTags 全量替换优秀作文标签，语义同 replaceArticleTags。
func replaceEssayTags(tx *gorm.DB, essayID uint, tagIDs []uint) error {
	tagIDs = validTagIDs(tx, dedupeTagIDs(tagIDs))
	if err := tx.Where("essay_id = ?", essayID).Delete(&model.EssayExampleTag{}).Error; err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	rows := make([]model.EssayExampleTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		rows = append(rows, model.EssayExampleTag{EssayID: essayID, TagID: tagID})
	}
	return tx.Create(&rows).Error
}

// ArticleDownloadSummary 后台文章列表附带的下载资料汇总（读时聚合，不落库）。
type ArticleDownloadSummary struct {
	Total      int `json:"total"`
	PublicFree int `json:"publicFree"`
}

// ArticleListItem 后台文章列表项：文章字段 + 关联标签（平铺 tags 便于表格绑定）。
type ArticleListItem struct {
	model.Article
	Tags            []model.Tag            `json:"tags"`
	DownloadSummary ArticleDownloadSummary `json:"downloadSummary"`
}

// articleTagRow 批量查询 article_tags JOIN tags 时的扫描行。
type articleTagRow struct {
	ArticleID uint `gorm:"column:article_id"`
	model.Tag
}

// articleTagsByArticleIDs 批量加载多篇文章的标签，按 sort_order DESC, id DESC。
func articleTagsByArticleIDs(articleIDs []uint) map[uint][]model.Tag {
	result := make(map[uint][]model.Tag, len(articleIDs))
	if len(articleIDs) == 0 {
		return result
	}
	var rows []articleTagRow
	database.DB.Table("article_tags").
		Select("article_tags.article_id, tags.*").
		Joins("JOIN tags ON tags.id = article_tags.tag_id").
		Where("article_tags.article_id IN ?", articleIDs).
		Order("tags.sort_order DESC, tags.id DESC").
		Scan(&rows)
	for _, row := range rows {
		tag := row.Tag
		result[row.ArticleID] = append(result[row.ArticleID], tag)
	}
	return result
}

// downloadSummaryByArticleIDs 批量统计文章关联下载资料总数及前台可下载数。
func downloadSummaryByArticleIDs(articleIDs []uint) map[uint]ArticleDownloadSummary {
	result := make(map[uint]ArticleDownloadSummary, len(articleIDs))
	if len(articleIDs) == 0 {
		return result
	}
	var files []model.DownloadFile
	database.DB.Model(&model.DownloadFile{}).
		Select("owner_id", "tags", "access_type", "status").
		Where("owner_type = ? AND owner_id IN ?", model.TargetTypeArticle, articleIDs).
		Find(&files)
	for _, f := range files {
		s := result[f.OwnerID]
		s.Total++
		if model.IsPublicFreeDownload(f) {
			s.PublicFree++
		}
		result[f.OwnerID] = s
	}
	return result
}

// BuildArticleListItems 将文章列表与标签、下载汇总映射组装为列表响应项。
func BuildArticleListItems(
	articles []model.Article,
	tagsByArticleID map[uint][]model.Tag,
	downloadsByArticleID map[uint]ArticleDownloadSummary,
) []ArticleListItem {
	items := make([]ArticleListItem, 0, len(articles))
	for _, article := range articles {
		tags := tagsByArticleID[article.ID]
		if tags == nil {
			tags = []model.Tag{}
		}
		summary := downloadsByArticleID[article.ID]
		items = append(items, ArticleListItem{
			Article:         article,
			Tags:            tags,
			DownloadSummary: summary,
		})
	}
	return items
}

// EssayListItem 后台优秀作文列表项：范文字段 + 关联标签（平铺 tags 便于表格绑定）。
type EssayListItem struct {
	model.EssayExample
	Tags []model.Tag `json:"tags"`
}

// essayTagRow 批量查询 essay_example_tags JOIN tags 时的扫描行。
type essayTagRow struct {
	EssayID uint `gorm:"column:essay_id"`
	model.Tag
}

// essayTagsByEssayIDs 批量加载多篇优秀作文的标签，按 sort_order DESC, id DESC。
func essayTagsByEssayIDs(essayIDs []uint) map[uint][]model.Tag {
	result := make(map[uint][]model.Tag, len(essayIDs))
	if len(essayIDs) == 0 {
		return result
	}
	var rows []essayTagRow
	database.DB.Table("essay_example_tags").
		Select("essay_example_tags.essay_id, tags.*").
		Joins("JOIN tags ON tags.id = essay_example_tags.tag_id").
		Where("essay_example_tags.essay_id IN ?", essayIDs).
		Order("tags.sort_order DESC, tags.id DESC").
		Scan(&rows)
	for _, row := range rows {
		tag := row.Tag
		result[row.EssayID] = append(result[row.EssayID], tag)
	}
	return result
}

// BuildEssayListItems 将优秀作文列表与标签映射组装为列表响应项。
func BuildEssayListItems(essays []model.EssayExample, tagsByEssayID map[uint][]model.Tag) []EssayListItem {
	items := make([]EssayListItem, 0, len(essays))
	for _, essay := range essays {
		tags := tagsByEssayID[essay.ID]
		if tags == nil {
			tags = []model.Tag{}
		}
		items = append(items, EssayListItem{EssayExample: essay, Tags: tags})
	}
	return items
}

// articleTags 关联表 join tags，按排序返回标签实体。
func (s *Server) articleTags(articleID uint) []model.Tag {
	var tags []model.Tag
	database.DB.Joins("JOIN article_tags ON article_tags.tag_id = tags.id").
		Where("article_tags.article_id = ?", articleID).
		Order("tags.sort_order DESC, tags.id DESC").
		Find(&tags)
	return tags
}

// essayTags 关联表 join tags，按排序返回标签实体。
func (s *Server) essayTags(essayID uint) []model.Tag {
	var tags []model.Tag
	database.DB.Joins("JOIN essay_example_tags ON essay_example_tags.tag_id = tags.id").
		Where("essay_example_tags.essay_id = ?", essayID).
		Order("tags.sort_order DESC, tags.id DESC").
		Find(&tags)
	return tags
}
