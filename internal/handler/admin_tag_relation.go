// Day5/Day6：后台内容标签关联（article_tags / essay_example_tags）。
// 提供「整体设置」与「查询」两类接口，设置采用先清后插的全量替换语义。
package handler

import (
	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type setTagsRequest struct {
	TagIDs []uint `json:"tagIds"`
}

// validTagIDs 过滤出实际存在的标签 ID，避免写入无效关联（本表无外键约束）。
func validTagIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return nil
	}
	var existing []uint
	database.DB.Model(&model.Tag{}).Where("id IN ?", ids).Pluck("id", &existing)
	return existing
}

// SetArticleTags 全量设置文章标签：事务内删除旧关联后插入去重后的新关联。
func (s *Server) SetArticleTags(c *gin.Context) {
	articleID, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var article model.Article
	if err := database.DB.First(&article, articleID).Error; err != nil {
		response.NotFound(c, "article not found")
		return
	}
	var req setTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tagIDs := validTagIDs(req.TagIDs)
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("article_id = ?", article.ID).Delete(&model.ArticleTag{}).Error; err != nil {
			return err
		}
		seen := map[uint]bool{}
		for _, tagID := range tagIDs {
			if seen[tagID] {
				continue
			}
			seen[tagID] = true
			if err := tx.Create(&model.ArticleTag{ArticleID: article.ID, TagID: tagID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, "failed to set article tags")
		return
	}
	response.Success(c, s.articleTags(article.ID))
}

// ListArticleTags 返回文章当前关联的标签详情列表。
func (s *Server) ListArticleTags(c *gin.Context) {
	articleID, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var article model.Article
	if err := database.DB.First(&article, articleID).Error; err != nil {
		response.NotFound(c, "article not found")
		return
	}
	response.Success(c, s.articleTags(article.ID))
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

// SetEssayTags 全量设置优秀作文标签，语义同 SetArticleTags。
func (s *Server) SetEssayTags(c *gin.Context) {
	essayID, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var essay model.EssayExample
	if err := database.DB.First(&essay, essayID).Error; err != nil {
		response.NotFound(c, "essay example not found")
		return
	}
	var req setTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tagIDs := validTagIDs(req.TagIDs)
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("essay_id = ?", essay.ID).Delete(&model.EssayExampleTag{}).Error; err != nil {
			return err
		}
		seen := map[uint]bool{}
		for _, tagID := range tagIDs {
			if seen[tagID] {
				continue
			}
			seen[tagID] = true
			if err := tx.Create(&model.EssayExampleTag{EssayID: essay.ID, TagID: tagID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, "failed to set essay tags")
		return
	}
	response.Success(c, s.essayTags(essay.ID))
}

// ListEssayTags 返回优秀作文当前关联的标签详情列表。
func (s *Server) ListEssayTags(c *gin.Context) {
	essayID, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var essay model.EssayExample
	if err := database.DB.First(&essay, essayID).Error; err != nil {
		response.NotFound(c, "essay example not found")
		return
	}
	response.Success(c, s.essayTags(essay.ID))
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
