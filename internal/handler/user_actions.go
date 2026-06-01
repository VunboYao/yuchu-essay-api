// Day8/Day10：需 UserAuth 的「我的」统计、收藏、浏览历史、下载记录。
package handler

import (
	"strconv"

	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type favoriteRequest struct {
	TargetType string `json:"targetType" binding:"required"`
	TargetID   uint   `json:"targetId" binding:"required"`
}

func (s *Server) ListFavorites(c *gin.Context) {
	userID := currentAuthID(c)
	var list []model.Favorite
	if err := database.DB.Where("user_id = ?", userID).Order("id DESC").Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list favorites")
		return
	}
	response.Success(c, list)
}

// AddFavorite 收藏文章或优秀作文；唯一索引去重，仅在真正新增时回写目标 favorite_count。
func (s *Server) AddFavorite(c *gin.Context) {
	userID := currentAuthID(c)
	var req favoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	fav := model.Favorite{UserID: userID, TargetType: req.TargetType, TargetID: req.TargetID}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&fav)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			return adjustFavoriteCount(tx, req.TargetType, req.TargetID, 1)
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, "failed to add favorite")
		return
	}
	response.Success(c, fav)
}

// RemoveFavorite 取消收藏；仅在确实删除记录时递减目标 favorite_count（下限 0）。
func (s *Server) RemoveFavorite(c *gin.Context) {
	userID := currentAuthID(c)
	targetType := c.Query("targetType")
	targetID, _ := strconv.Atoi(c.Query("targetId"))
	if targetType == "" || targetID == 0 {
		response.BadRequest(c, "targetType and targetId are required")
		return
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
			Delete(&model.Favorite{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			return adjustFavoriteCount(tx, targetType, uint(targetID), -1)
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, "failed to remove favorite")
		return
	}
	response.Success(c, gin.H{"targetType": targetType, "targetId": targetID})
}

// adjustFavoriteCount 按目标类型回写收藏计数；delta 为负时以 0 兜底，未知类型不处理。
func adjustFavoriteCount(tx *gorm.DB, targetType string, targetID uint, delta int) error {
	var target interface{}
	switch targetType {
	case model.TargetTypeArticle:
		target = &model.Article{}
	case model.TargetTypeEssay:
		target = &model.EssayExample{}
	default:
		return nil
	}
	expr := gorm.Expr("favorite_count + ?", delta)
	if delta < 0 {
		expr = gorm.Expr("GREATEST(favorite_count + ?, 0)", delta)
	}
	return tx.Model(target).Where("id = ?", targetID).UpdateColumn("favorite_count", expr).Error
}

func (s *Server) History(c *gin.Context) {
	userID := currentAuthID(c)
	var list []model.ViewLog
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(100).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list history")
		return
	}
	response.Success(c, list)
}

func (s *Server) Downloads(c *gin.Context) {
	userID := currentAuthID(c)
	var list []model.DownloadLog
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(100).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list downloads")
		return
	}
	response.Success(c, list)
}

// Mine 我的页概览：收藏数、浏览记录数、下载记录数。
func (s *Server) Mine(c *gin.Context) {
	userID := currentAuthID(c)
	var favCount int64
	var viewCount int64
	var downloadCount int64
	database.DB.Model(&model.Favorite{}).Where("user_id = ?", userID).Count(&favCount)
	database.DB.Model(&model.ViewLog{}).Where("user_id = ?", userID).Count(&viewCount)
	database.DB.Model(&model.DownloadLog{}).Where("user_id = ?", userID).Count(&downloadCount)
	response.Success(c, gin.H{
		"favoriteCount": favCount,
		"viewCount":     viewCount,
		"downloadCount": downloadCount,
	})
}
