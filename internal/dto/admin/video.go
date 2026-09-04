// 视频号讲解视频 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package admin

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"course_server/internal/dto"
	"course_server/internal/model"
)

const (
	maxVideoTitleRunes = 120
	maxVideoFeedIDLen  = 255
)

// VideoWriteRequest 讲解视频创建/更新共用入参。
type VideoWriteRequest struct {
	Title     string `json:"title" binding:"required"`
	FeedID    string `json:"feedId" binding:"required"`
	Grade     int    `json:"grade"`
	Semester  string `json:"semester"`
	TagIDs    []uint `json:"tagIds"`
	ArticleID *uint  `json:"articleId"`
	EssayID   *uint  `json:"essayId"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

// CreateVideoRequest 创建讲解视频。
type CreateVideoRequest = VideoWriteRequest

// UpdateVideoRequest 更新讲解视频。
type UpdateVideoRequest = VideoWriteRequest

func optionalUint(p *uint) *uint {
	if p == nil || *p == 0 {
		return nil
	}
	return p
}

func (r VideoWriteRequest) normalize() (model.Video, error) {
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return model.Video{}, fmt.Errorf("title is required")
	}
	if utf8.RuneCountInString(title) > maxVideoTitleRunes {
		return model.Video{}, fmt.Errorf("title exceeds max %d", maxVideoTitleRunes)
	}
	feedID := strings.TrimSpace(r.FeedID)
	if feedID == "" {
		return model.Video{}, fmt.Errorf("feedId is required")
	}
	if len(feedID) > maxVideoFeedIDLen {
		return model.Video{}, fmt.Errorf("feedId exceeds max %d", maxVideoFeedIDLen)
	}
	// grade 合法值是 0（不限）或 1–6；int 不传与传 0 相同
	if r.Grade != 0 && (r.Grade < 1 || r.Grade > 6) {
		return model.Video{}, fmt.Errorf("grade must be 0 or 1-6")
	}
	semester := strings.TrimSpace(r.Semester)
	if semester != "" && !model.ValidSemester(semester) {
		return model.Video{}, fmt.Errorf("invalid semester")
	}
	return model.Video{
		Title:     title,
		FeedID:    feedID,
		Grade:     r.Grade,
		Semester:  semester,
		ArticleID: optionalUint(r.ArticleID),
		EssayID:   optionalUint(r.EssayID),
		SortOrder: r.SortOrder,
	}, nil
}

func (r VideoWriteRequest) ValidateForCreate() (model.Video, error) {
	item, err := r.normalize()
	if err != nil {
		return model.Video{}, err
	}
	status, err := dto.ResolveStatus(r.Status, model.StatusEnabled)
	if err != nil {
		return model.Video{}, err
	}
	item.Status = status
	return item, nil
}

func (r VideoWriteRequest) ValidateForUpdate(existing model.Video) (model.Video, error) {
	item, err := r.normalize()
	if err != nil {
		return model.Video{}, err
	}
	status, err := dto.ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return model.Video{}, err
	}
	item.Status = status
	return item, nil
}
