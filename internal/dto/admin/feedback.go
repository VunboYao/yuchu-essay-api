package admin

import (
	"fmt"
	"strings"
)

// ReplyFeedbackRequest 管理员回复意见反馈。
type ReplyFeedbackRequest struct {
	Reply string `json:"reply" binding:"required"`
}

func (r ReplyFeedbackRequest) NormalizeReply() (string, error) {
	reply := strings.TrimSpace(r.Reply)
	if reply == "" {
		return "", fmt.Errorf("reply is required")
	}
	if len([]rune(reply)) > 1000 {
		return "", fmt.Errorf("reply too long")
	}
	return reply, nil
}
