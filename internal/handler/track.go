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

// 埋点入参不含 userId，由 OptionalUserAuth + currentAuthID 服务端赋值（未登录为 0）。
type trackViewRequest struct {
	TargetType  string `json:"targetType"`
	TargetID    uint   `json:"targetId"`
	PagePath    string `json:"pagePath"`
	SourceScene string `json:"sourceScene"`
}

type trackShareRequest struct {
	TargetType   string `json:"targetType"`
	TargetID     uint   `json:"targetId"`
	ShareChannel string `json:"shareChannel"`
	SourceScene  string `json:"sourceScene"`
}

type trackSearchRequest struct {
	Keyword     string `json:"keyword"`
	ResultCount int    `json:"resultCount"`
	SourceScene string `json:"sourceScene"`
}

type trackAdRequest struct {
	SlotKey    string `json:"slotKey" binding:"required"`
	EventType  string `json:"eventType" binding:"required"`
	PagePath   string `json:"pagePath"`
	TargetType string `json:"targetType"`
	TargetID   uint   `json:"targetId"`
}

var errTrackTargetPairRequired = errors.New("targetType and targetId must be provided together")

// validateContentTarget 埋点关联内容：target 须成对；有 target 时须为已发布文章或启用范文。
func validateContentTarget(db *gorm.DB, targetType string, targetID uint) error {
	targetType = strings.TrimSpace(targetType)
	if targetID > 0 || targetType != "" {
		if targetType == "" || targetID == 0 {
			return errTrackTargetPairRequired
		}
		return model.CheckFavoriteTargetExists(db, targetType, targetID)
	}
	return nil
}

func respondContentTargetError(c *gin.Context, err error, serverMsg string) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, errTrackTargetPairRequired):
		response.BadRequest(c, err.Error())
	case errors.Is(err, model.ErrInvalidFavoriteTargetType):
		response.BadRequest(c, err.Error())
	case errors.Is(err, model.ErrFavoriteArticleNotFound), errors.Is(err, model.ErrFavoriteEssayNotFound):
		response.NotFound(c, err.Error())
	default:
		response.ServerError(c, serverMsg)
	}
	return true
}

// TrackView 记录浏览；userId 仅来自 Bearer JWT（未登录为 0），不接受 body 传入。
func (s *Server) TrackView(c *gin.Context) {
	var req trackViewRequest
	if !bindJSON(c, &req) {
		return
	}
	targetType := strings.TrimSpace(req.TargetType)
	targetID := req.TargetID
	if err := validateContentTarget(database.DB, targetType, targetID); respondContentTargetError(c, err, "failed to track view") {
		return
	}
	log := model.ViewLog{
		UserID:      currentAuthID(c),
		TargetType:  targetType,
		TargetID:    targetID,
		PagePath:    req.PagePath,
		SourceScene: req.SourceScene,
	}
	if err := database.DB.Create(&log).Error; err != nil {
		response.ServerError(c, "failed to track view")
		return
	}
	response.Success(c, log)
}

// TrackShare 记录分享事件；目标为已发布文章时事务内同步递增 share_count。
func (s *Server) TrackShare(c *gin.Context) {
	var req trackShareRequest
	if !bindJSON(c, &req) {
		return
	}
	targetType := strings.TrimSpace(req.TargetType)
	targetID := req.TargetID
	log := model.ShareLog{
		UserID:       currentAuthID(c),
		TargetType:   targetType,
		TargetID:     targetID,
		ShareChannel: req.ShareChannel,
		SourceScene:  req.SourceScene,
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := validateContentTarget(tx, targetType, targetID); err != nil {
			return err
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		if log.TargetType == model.TargetTypeArticle && log.TargetID > 0 {
			return tx.Model(&model.Article{}).Where("id = ?", log.TargetID).
				UpdateColumn("share_count", gorm.Expr("share_count + ?", 1)).Error
		}
		return nil
	})
	if respondContentTargetError(c, err, "failed to track share") {
		return
	}
	response.Success(c, log)
}

func (s *Server) TrackSearch(c *gin.Context) {
	var req trackSearchRequest
	if !bindJSON(c, &req) {
		return
	}
	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		response.BadRequest(c, "keyword is required")
		return
	}
	log := model.SearchLog{
		UserID:      currentAuthID(c),
		Keyword:     keyword,
		ResultCount: req.ResultCount,
		SourceScene: req.SourceScene,
	}
	if err := database.DB.Create(&log).Error; err != nil {
		response.ServerError(c, "failed to track search")
		return
	}
	response.Success(c, log)
}

func (s *Server) TrackAd(c *gin.Context) {
	var req trackAdRequest
	if !bindJSON(c, &req) {
		return
	}
	slotKey := strings.TrimSpace(req.SlotKey)
	if slotKey == "" {
		response.BadRequest(c, "slotKey is required")
		return
	}
	eventType := strings.TrimSpace(req.EventType)
	if !model.ValidAdEventType(eventType) {
		response.BadRequest(c, "eventType must be expose or click")
		return
	}
	if eventType == model.AdEventRewardedComplete {
		response.BadRequest(c, "use POST /download-files/:id/ad-unlock for rewarded video unlock")
		return
	}
	if _, err := model.FindAdSlotByKey(database.DB, slotKey); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.BadRequest(c, "unknown ad slot")
			return
		}
		response.ServerError(c, "failed to find ad slot")
		return
	}
	targetType := strings.TrimSpace(req.TargetType)
	targetID := req.TargetID
	if err := validateContentTarget(database.DB, targetType, targetID); respondContentTargetError(c, err, "failed to track ad") {
		return
	}
	log := model.AdEvent{
		UserID:     currentAuthID(c),
		SlotKey:    slotKey,
		EventType:  eventType,
		PagePath:   strings.TrimSpace(req.PagePath),
		TargetType: targetType,
		TargetID:   targetID,
	}
	if err := database.DB.Create(&log).Error; err != nil {
		response.ServerError(c, "failed to record ad event")
		return
	}
	response.Success(c, log)
}
