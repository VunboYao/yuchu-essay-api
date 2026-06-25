// 优秀作文列表筛选（后台与公开接口共用）。
package handler

import (
	"strings"

	"course_server/internal/database"
	"course_server/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// essayBaseFilters 年级、关联文章筛选（不含标签/关键词，供 filterTags 子查询复用）。
func essayBaseFilters(db *gorm.DB, c *gin.Context) *gorm.DB {
	if grade := strings.TrimSpace(c.Query("grade")); grade != "" {
		db = db.Where("grade = ?", grade)
	}
	if articleID := strings.TrimSpace(c.Query("articleId")); articleID != "" {
		db = db.Where("article_id = ?", articleID)
	}
	return db
}

// essayFilters 优秀作文列表共用的年级、关联文章、标签、关键词筛选。
func essayFilters(db *gorm.DB, c *gin.Context) *gorm.DB {
	db = essayBaseFilters(db, c)
	if tagID := strings.TrimSpace(c.Query("tagId")); tagID != "" {
		db = db.Where(
			"id IN (SELECT essay_id FROM essay_example_tags WHERE tag_id = ?)",
			tagID,
		)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		db = db.Where("title LIKE ? OR body LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	return db
}

// listPublicEssayFilterTags 返回当前年级/文章范围内可用于筛选的启用标签。
func listPublicEssayFilterTags(c *gin.Context) []model.Tag {
	subQuery := essayBaseFilters(database.DB.Model(&model.EssayExample{}), c).
		Where("status = ?", model.StatusEnabled).
		Select("essay_examples.id")
	var tags []model.Tag
	database.DB.Model(&model.Tag{}).
		Joins("JOIN essay_example_tags ON essay_example_tags.tag_id = tags.id").
		Where("essay_example_tags.essay_id IN (?) AND tags.status = ?", subQuery, model.StatusEnabled).
		Order("tags.sort_order DESC, tags.id DESC").
		Distinct("tags.*").
		Find(&tags)
	if tags == nil {
		return []model.Tag{}
	}
	return tags
}
