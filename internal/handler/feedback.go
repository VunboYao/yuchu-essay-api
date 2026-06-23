// 意见反馈：小程序提交/查询与后台回复。
package handler

import (
	"strings"
	"unicode/utf8"

	"course_server/internal/database"
	"course_server/internal/dto/admin"
	"course_server/internal/model"
	"course_server/internal/pkg/jsontime"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createFeedbackRequest struct {
	Content string `json:"content" binding:"required"`
}

// FeedbackListItem 反馈列表项（小程序与后台共用字段子集）。
type FeedbackListItem struct {
	ID          uint           `json:"id"`
	UserID      uint           `json:"userId"`
	UserNickname string        `json:"userNickname,omitempty"`
	Content     string         `json:"content"`
	Status      string         `json:"status"`
	Reply       string         `json:"reply,omitempty"`
	RepliedAt   *jsontime.Time `json:"repliedAt,omitempty"`
	CreatedAt   jsontime.Time  `json:"createdAt"`
}

// CreateFeedback 提交意见反馈。
func (s *Server) CreateFeedback(c *gin.Context) {
	userID := currentAuthID(c)
	var req createFeedbackRequest
	if !bindJSON(c, &req) {
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		response.BadRequest(c, "content is required")
		return
	}
	if utf8.RuneCountInString(content) > 300 {
		response.BadRequest(c, "content too long")
		return
	}
	item := model.Feedback{
		UserID:  userID,
		Content: content,
		Status:  model.FeedbackStatusPending,
	}
	if err := database.DB.Create(&item).Error; err != nil {
		response.ServerError(c, "failed to create feedback")
		return
	}
	response.Success(c, toFeedbackListItem(item, ""))
}

// ListUserFeedbacks 当前用户反馈列表。
func (s *Server) ListUserFeedbacks(c *gin.Context) {
	userID := currentAuthID(c)
	page, pageSize := pageParams(c)
	list, total, err := loadFeedbackListPage(userID, "", "", page, pageSize)
	if err != nil {
		response.ServerError(c, "failed to list feedbacks")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

// ListFeedbacks 后台分页查询反馈，支持 status 与关键词筛选。
func (s *Server) ListFeedbacks(c *gin.Context) {
	page, pageSize := pageParams(c)
	status := strings.TrimSpace(c.Query("status"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	list, total, err := loadFeedbackListPage(0, status, keyword, page, pageSize)
	if err != nil {
		response.ServerError(c, "failed to list feedbacks")
		return
	}
	response.Page(c, list, total, page, pageSize)
}

// ReplyFeedback 管理员回复反馈。
func (s *Server) ReplyFeedback(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req admin.ReplyFeedbackRequest
	if !bindJSON(c, &req) {
		return
	}
	reply, err := req.NormalizeReply()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	var item model.Feedback
	if err := database.DB.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "feedback not found")
			return
		}
		response.ServerError(c, "failed to load feedback")
		return
	}
	now := jsontime.Now()
	adminID := currentAuthID(c)
	updates := map[string]interface{}{
		"reply":      reply,
		"status":     model.FeedbackStatusReplied,
		"replied_at": &now,
		"replied_by": adminID,
	}
	if err := database.DB.Model(&item).Updates(updates).Error; err != nil {
		response.ServerError(c, "failed to reply feedback")
		return
	}
	if err := database.DB.First(&item, id).Error; err != nil {
		response.ServerError(c, "failed to load feedback")
		return
	}
	nickname := loadUserNickname(item.UserID)
	response.Success(c, toFeedbackListItem(item, nickname))
}

func loadFeedbackListPage(userID uint, status, keyword string, page, pageSize int) ([]FeedbackListItem, int64, error) {
	db := database.DB.Model(&model.Feedback{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("content LIKE ? OR reply LIKE ?", like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Feedback
	if err := paginate(db.Order("id DESC"), page, pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	nicknames := loadUserNicknames(collectFeedbackUserIDs(rows))
	items := make([]FeedbackListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toFeedbackListItem(row, nicknames[row.UserID]))
	}
	return items, total, nil
}

func toFeedbackListItem(row model.Feedback, nickname string) FeedbackListItem {
	return FeedbackListItem{
		ID:           row.ID,
		UserID:       row.UserID,
		UserNickname: nickname,
		Content:      row.Content,
		Status:       row.Status,
		Reply:        row.Reply,
		RepliedAt:    row.RepliedAt,
		CreatedAt:    row.CreatedAt,
	}
}

func collectFeedbackUserIDs(rows []model.Feedback) []uint {
	ids := make([]uint, 0, len(rows))
	seen := map[uint]bool{}
	for _, row := range rows {
		if seen[row.UserID] {
			continue
		}
		seen[row.UserID] = true
		ids = append(ids, row.UserID)
	}
	return ids
}

func loadUserNicknames(userIDs []uint) map[uint]string {
	out := map[uint]string{}
	if len(userIDs) == 0 {
		return out
	}
	var users []model.User
	if err := database.DB.Select("id", "nickname").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return out
	}
	for _, user := range users {
		out[user.ID] = user.Nickname
	}
	return out
}

func loadUserNickname(userID uint) string {
	return loadUserNicknames([]uint{userID})[userID]
}
