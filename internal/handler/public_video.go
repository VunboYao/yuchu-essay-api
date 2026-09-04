// 小程序公开讲解视频列表：仅上架项，按年级/册次/标签筛选。
package handler

import (
	"strconv"
	"strings"

	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const publicVideoListLimit = 200

// PublicVideoItem 公开列表单项：视频字段 + 可选关联卡片。
type PublicVideoItem struct {
	ID       uint                   `json:"id"`
	Title    string                 `json:"title"`
	FeedID   string                 `json:"feedId"`
	Grade    int                    `json:"grade"`
	Semester string                 `json:"semester"`
	Article  *ContentArticleSummary `json:"article"`
	Essay    *ContentEssaySummary   `json:"essay"`
}

// ListPublicVideos 公开讲解视频列表。finderUserName 放响应顶层一次。
func (s *Server) ListPublicVideos(c *gin.Context) {
	q := database.DB.Model(&model.Video{}).Where("status = ?", model.StatusEnabled)
	if grade, _ := strconv.Atoi(strings.TrimSpace(c.Query("grade"))); grade > 0 {
		q = q.Where("grade = ? OR grade = 0", grade)
	}
	if semester := strings.TrimSpace(c.Query("semester")); semester != "" {
		q = q.Where("semester = ? OR semester = ''", semester)
	}
	if tagID, _ := strconv.Atoi(strings.TrimSpace(c.Query("tagId"))); tagID > 0 {
		q = q.Joins("JOIN video_tags ON video_tags.video_id = videos.id AND video_tags.tag_id = ?", tagID)
	}

	var rows []model.Video
	if err := q.Order("sort_order DESC, id DESC").Limit(publicVideoListLimit).Find(&rows).Error; err != nil {
		response.ServerError(c, "failed to list videos")
		return
	}

	finderUserName, err := model.GetSystemSetting(database.DB, model.SettingKeyChannelsFinderUserName)
	if err != nil {
		response.ServerError(c, "failed to get video channel")
		return
	}

	articleIDs, essayIDs := collectVideoContentIDs(rows)
	articles, essays, linkedArticles, err := loadContentTargets(database.DB, articleIDs, essayIDs)
	if err != nil {
		response.ServerError(c, "failed to load video targets")
		return
	}

	response.Success(c, gin.H{
		"finderUserName": finderUserName,
		"list":           BuildPublicVideoItems(rows, articles, essays, linkedArticles, s.mediaResolver()),
	})
}

func collectVideoContentIDs(rows []model.Video) (articleIDs, essayIDs []uint) {
	seenArticle := map[uint]bool{}
	seenEssay := map[uint]bool{}
	for _, row := range rows {
		if row.ArticleID != nil && *row.ArticleID > 0 && !seenArticle[*row.ArticleID] {
			seenArticle[*row.ArticleID] = true
			articleIDs = append(articleIDs, *row.ArticleID)
		}
		if row.EssayID != nil && *row.EssayID > 0 && !seenEssay[*row.EssayID] {
			seenEssay[*row.EssayID] = true
			essayIDs = append(essayIDs, *row.EssayID)
		}
	}
	return articleIDs, essayIDs
}

// BuildPublicVideoItems 组装公开列表项：未发布/未启用的关联内容出 null，封面走 OSS 解析。
func BuildPublicVideoItems(
	rows []model.Video,
	articles map[uint]model.Article,
	essays map[uint]model.EssayExample,
	linkedArticles map[uint]ArticleTitleRef,
	resolve MediaURLResolver,
) []PublicVideoItem {
	items := make([]PublicVideoItem, 0, len(rows))
	for _, row := range rows {
		item := PublicVideoItem{
			ID:       row.ID,
			Title:    row.Title,
			FeedID:   row.FeedID,
			Grade:    row.Grade,
			Semester: row.Semester,
		}
		if row.ArticleID != nil && *row.ArticleID > 0 {
			ok, target := buildContentTarget(model.TargetTypeArticle, *row.ArticleID, articles, essays, linkedArticles)
			if ok {
				if summary, ok := resolveContentArticleTarget(target, resolve).(ContentArticleSummary); ok {
					item.Article = &summary
				}
			}
		}
		if row.EssayID != nil && *row.EssayID > 0 {
			ok, target := buildContentTarget(model.TargetTypeEssay, *row.EssayID, articles, essays, linkedArticles)
			if ok {
				if summary, ok := target.(ContentEssaySummary); ok {
					item.Essay = &summary
				}
			}
		}
		items = append(items, item)
	}
	return items
}
