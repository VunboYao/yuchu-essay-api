// Day11：小程序行为埋点（公开接口），写入 view/share/search/ad 日志表。
package handler

import (
	"errors"
	"strings"

	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TrackView 记录浏览；body 未带 userId 时尝试从 Bearer token 解析（未登录则为 0）。
func (s *Server) TrackView(c *gin.Context) {
	var req model.ViewLog
	if !bindJSON(c, &req) {
		return
	}
	if req.UserID == 0 {
		req.UserID = currentAuthID(c)
	}
	if err := database.DB.Create(&req).Error; err != nil {
		response.ServerError(c, "failed to track view")
		return
	}
	response.Success(c, req)
}

// TrackShare 记录分享事件；目标为文章时事务内同步递增 share_count。
func (s *Server) TrackShare(c *gin.Context) {
	var req model.ShareLog
	if !bindJSON(c, &req) {
		return
	}
	if req.UserID == 0 {
		req.UserID = currentAuthID(c)
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&req).Error; err != nil {
			return err
		}
		if req.TargetType == model.TargetTypeArticle && req.TargetID > 0 {
			return tx.Model(&model.Article{}).Where("id = ?", req.TargetID).
				UpdateColumn("share_count", gorm.Expr("share_count + ?", 1)).Error
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, "failed to track share")
		return
	}
	response.Success(c, req)
}

func (s *Server) TrackSearch(c *gin.Context) {
	var req model.SearchLog
	if !bindJSON(c, &req) {
		return
	}
	if req.UserID == 0 {
		req.UserID = currentAuthID(c)
	}
	if err := database.DB.Create(&req).Error; err != nil {
		response.ServerError(c, "failed to track search")
		return
	}
	response.Success(c, req)
}

func (s *Server) TrackAd(c *gin.Context) {
	var req model.AdEvent
	if !bindJSON(c, &req) {
		return
	}
	slotKey := strings.TrimSpace(req.SlotKey)
	if slotKey == "" {
		response.BadRequest(c, "slotKey is required")
		return
	}
	if !model.ValidAdEventType(strings.TrimSpace(req.EventType)) {
		response.BadRequest(c, "eventType must be expose, click or rewarded_complete")
		return
	}
	if req.EventType == model.AdEventRewardedComplete {
		response.BadRequest(c, "use POST /download-files/:id/ad-unlock for rewarded video unlock")
		return
	}
	if _, err := model.FindAdSlotByKey(database.DB, slotKey); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.BadRequest(c, "unknown ad slot")
			return
		}
		response.ServerError(c, "failed to track ad")
		return
	}
	req.SlotKey = slotKey
	req.EventType = strings.TrimSpace(req.EventType)
	if req.UserID == 0 {
		req.UserID = currentAuthID(c)
	}
	if err := database.DB.Create(&req).Error; err != nil {
		response.ServerError(c, "failed to track ad")
		return
	}
	response.Success(c, req)
}
