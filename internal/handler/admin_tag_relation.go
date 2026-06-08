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
