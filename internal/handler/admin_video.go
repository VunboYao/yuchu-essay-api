// 后台视频号讲解视频 CRUD，以及全站视频号 ID（system_settings 固定键）。
package handler

import (
	"errors"
	"strings"

	"course_server/internal/database"
	admindto "course_server/internal/dto/admin"
	"course_server/internal/middleware"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListVideos 后台列表（含每行 tags；筛 grade / semester / status / keyword）。
func (s *Server) ListVideos(c *gin.Context) {
	page, pageSize := pageParams(c)
	db := videoFilters(database.DB.Model(&model.Video{}), c)
	var total int64
	db.Count(&total)
	var list []model.Video
	if err := paginate(db.Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list videos")
		return
	}
	videoIDs := make([]uint, 0, len(list))
	for _, item := range list {
		videoIDs = append(videoIDs, item.ID)
	}
	items := BuildVideoListItems(list, videoTagsByVideoIDs(videoIDs))
	response.Page(c, items, total, page, pageSize)
}

func videoFilters(db *gorm.DB, c *gin.Context) *gorm.DB {
	if grade := c.Query("grade"); grade != "" {
		db = db.Where("grade = ?", grade)
	}
	if semester := c.Query("semester"); semester != "" {
		db = db.Where("semester = ?", semester)
	}
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		db = db.Where("title LIKE ?", "%"+keyword+"%")
	}
	return db
}

// GetVideo 获取讲解视频详情（含关联标签）。
func (s *Server) GetVideo(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	item, findErr := model.FindVideoByID(database.DB, uint(id))
	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			response.NotFound(c, "video not found")
			return
		}
		response.ServerError(c, "failed to find video")
		return
	}
	response.Success(c, gin.H{
		"video": item,
		"tags":  s.videoTags(item.ID),
	})
}

// CreateVideo 创建讲解视频；feed-id 冲突返回 400。
func (s *Server) CreateVideo(c *gin.Context) {
	var req admindto.CreateVideoRequest
	if !bindJSON(c, &req) {
		return
	}
	video, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if !s.prepareVideoWrite(c, &video, 0) {
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := CreateAndReload(tx, &video, video.UpdateMap()); err != nil {
			return err
		}
		if req.TagIDs != nil {
			return replaceVideoTags(tx, video.ID, req.TagIDs)
		}
		return nil
	}); err != nil {
		response.ServerError(c, "failed to create video")
		return
	}
	middleware.SetOpLogRemark(c, "创建讲解视频：《"+video.Title+"》")
	middleware.SetOpLogTargetID(c, video.ID)
	response.Success(c, gin.H{
		"video": video,
		"tags":  s.videoTags(video.ID),
	})
}

// UpdateVideo 更新讲解视频。
func (s *Server) UpdateVideo(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	item, findErr := model.FindVideoByID(database.DB, uint(id))
	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			response.NotFound(c, "video not found")
			return
		}
		response.ServerError(c, "failed to find video")
		return
	}
	var req admindto.UpdateVideoRequest
	if !bindJSON(c, &req) {
		return
	}
	video, err := req.ValidateForUpdate(item)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if !s.prepareVideoWrite(c, &video, item.ID) {
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := UpdateAndReload(tx, &item, video.UpdateMap()); err != nil {
			return err
		}
		if req.TagIDs != nil {
			return replaceVideoTags(tx, item.ID, req.TagIDs)
		}
		return nil
	}); err != nil {
		response.ServerError(c, "failed to update video")
		return
	}
	middleware.SetOpLogRemark(c, "更新讲解视频：《"+item.Title+"》")
	response.Success(c, gin.H{
		"video": item,
		"tags":  s.videoTags(item.ID),
	})
}

// prepareVideoWrite 查重 feed-id 并校验关联专题/作文状态；失败时已写响应。
func (s *Server) prepareVideoWrite(c *gin.Context, video *model.Video, excludeID uint) bool {
	taken, err := model.FeedIDTaken(database.DB, video.FeedID, excludeID)
	if err != nil {
		response.ServerError(c, "failed to check feed-id")
		return false
	}
	if taken {
		response.BadRequest(c, "feed-id 已存在")
		return false
	}
	if err := model.CheckVideoArticleLink(database.DB, video.ArticleID); err != nil {
		response.BadRequest(c, err.Error())
		return false
	}
	if err := model.CheckVideoEssayLink(database.DB, video.EssayID); err != nil {
		response.BadRequest(c, err.Error())
		return false
	}
	return true
}

// DeleteVideo 物理删除讲解视频；须已下架，同时清掉 video_tags。
func (s *Server) DeleteVideo(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	item, findErr := model.FindVideoByID(database.DB, uint(id))
	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			response.NotFound(c, "video not found")
			return
		}
		response.ServerError(c, "failed to find video")
		return
	}
	if err := model.CheckVideoDeletable(item); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("video_id = ?", item.ID).Delete(&model.VideoTag{}).Error; err != nil {
			return err
		}
		return tx.Delete(&item).Error
	}); err != nil {
		response.ServerError(c, "failed to delete video")
		return
	}
	middleware.SetOpLogRemark(c, "删除讲解视频：《"+item.Title+"》")
	response.Success(c, gin.H{"id": id})
}

// GetVideoChannel 读取全站视频号 ID。
func (s *Server) GetVideoChannel(c *gin.Context) {
	val, err := model.GetSystemSetting(database.DB, model.SettingKeyChannelsFinderUserName)
	if err != nil {
		response.ServerError(c, "failed to get video channel")
		return
	}
	response.Success(c, gin.H{"finderUserName": val})
}

// UpdateVideoChannel 写入全站视频号 ID；空串视为清空。
func (s *Server) UpdateVideoChannel(c *gin.Context) {
	var req struct {
		FinderUserName string `json:"finderUserName"`
	}
	if !bindJSON(c, &req) {
		return
	}
	val := strings.TrimSpace(req.FinderUserName)
	if err := model.UpsertSystemSetting(database.DB, model.SettingKeyChannelsFinderUserName, val, "微信视频号 finder-user-name"); err != nil {
		response.ServerError(c, "failed to save video channel")
		return
	}
	middleware.SetOpLogRemark(c, "更新视频号 ID")
	response.Success(c, gin.H{"finderUserName": val})
}
