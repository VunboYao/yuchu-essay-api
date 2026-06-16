// Day11：后台 Dashboard 统计（需 AdminAuth）。
// 「今日」统计按服务器本地时区当日 00:00:00 起算（time.Local）。
package handler

import (
	"time"

	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type statTodayTotal struct {
	Today int64 `json:"today"`
	Total int64 `json:"total"`
}

// StatsOverview Dashboard 核心指标：今日/累计与分类计数（PV 取自 view_logs 埋点表）。
func (s *Server) StatsOverview(c *gin.Context) {
	usersToday, usersTotal := countTodayTotal(&model.User{})
	pvToday, pvTotal := countTodayTotal(&model.ViewLog{})
	todayStart := startOfToday()
	uvToday, err := countDistinctViewUsers(&todayStart)
	if err != nil {
		response.ServerError(c, "failed to load stats overview")
		return
	}
	uvTotal, err := countDistinctViewUsers(nil)
	if err != nil {
		response.ServerError(c, "failed to load stats overview")
		return
	}
	var articlesTotal int64
	var articlesPublished int64
	database.DB.Model(&model.Article{}).Count(&articlesTotal)
	database.DB.Model(&model.Article{}).Where("status = ?", model.ArticleStatusPublished).Count(&articlesPublished)
	var essaysTotal int64
	database.DB.Model(&model.EssayExample{}).Count(&essaysTotal)
	downloadsToday, downloadsTotal := countTodayTotal(&model.DownloadLog{})
	searchesToday, searchesTotal := countTodayTotal(&model.SearchLog{})
	favoritesToday, favoritesTotal := countTodayTotal(&model.Favorite{})
	sharesToday, sharesTotal := countTodayTotal(&model.ShareLog{})
	exposeToday, exposeTotal := countAdEventTodayTotal(model.AdEventExpose)
	clickToday, clickTotal := countAdEventTodayTotal(model.AdEventClick)
	rewardedToday, rewardedTotal := countAdEventTodayTotal(model.AdEventRewardedComplete)

	response.Success(c, gin.H{
		"users": statTodayTotal{Today: usersToday, Total: usersTotal},
		"pv":    statTodayTotal{Today: pvToday, Total: pvTotal},
		"uv":    statTodayTotal{Today: uvToday, Total: uvTotal},
		"articles": gin.H{
			"total":     articlesTotal,
			"published": articlesPublished,
		},
		"essays":    gin.H{"total": essaysTotal},
		"downloads": statTodayTotal{Today: downloadsToday, Total: downloadsTotal},
		"searches":  statTodayTotal{Today: searchesToday, Total: searchesTotal},
		"favorites": statTodayTotal{Today: favoritesToday, Total: favoritesTotal},
		"shares":    statTodayTotal{Today: sharesToday, Total: sharesTotal},
		"ads": gin.H{
			"expose":           statTodayTotal{Today: exposeToday, Total: exposeTotal},
			"click":            statTodayTotal{Today: clickToday, Total: clickTotal},
			"rewardedComplete": statTodayTotal{Today: rewardedToday, Total: rewardedTotal},
		},
	})
}

func (s *Server) StatsHotArticles(c *gin.Context) {
	var list []model.Article
	if err := database.DB.Omit("body_images").Order("view_count DESC, id DESC").Limit(10).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list hot articles")
		return
	}
	response.Success(c, list)
}

// StatsHotSearches 按关键词聚合 search_logs，取 Top 20。
func (s *Server) StatsHotSearches(c *gin.Context) {
	type row struct {
		Keyword string `json:"keyword"`
		Count   int64  `json:"count"`
	}
	var rows []row
	if err := database.DB.Model(&model.SearchLog{}).
		Select("keyword, count(*) as count").
		Group("keyword").
		Order("count DESC").
		Limit(20).
		Scan(&rows).Error; err != nil {
		response.ServerError(c, "failed to list hot searches")
		return
	}
	response.Success(c, rows)
}

// StatsHotDownloads 免费资料下载排行，按 download_files.download_count Top 20。
func (s *Server) StatsHotDownloads(c *gin.Context) {
	type row struct {
		FileID        uint   `json:"fileId"`
		FileName      string `json:"fileName"`
		FileType      string `json:"fileType"`
		OwnerType     string `json:"ownerType"`
		OwnerID       uint   `json:"ownerId"`
		DownloadCount int64  `json:"downloadCount"`
	}
	var rows []row
	if err := database.DB.Model(&model.DownloadFile{}).
		Select("id as file_id", "file_name", "file_type", "owner_type", "owner_id", "download_count").
		Where("status = ? AND access_type = ?", model.StatusEnabled, model.AccessTypeFree).
		Order("download_count DESC, id DESC").
		Limit(20).
		Scan(&rows).Error; err != nil {
		response.ServerError(c, "failed to list hot downloads")
		return
	}
	response.Success(c, rows)
}

// StatsAdEvents 广告曝光/点击/激励完成统计：全局 summary 与按 slotKey 拆分（累计）。
func (s *Server) StatsAdEvents(c *gin.Context) {
	// 今日曝光
	exposeToday, exposeTotal := countAdEventTodayTotal(model.AdEventExpose)
	// 今日点击
	clickToday, clickTotal := countAdEventTodayTotal(model.AdEventClick)
	// 今日激励完成
	rewardedToday, rewardedTotal := countAdEventTodayTotal(model.AdEventRewardedComplete)

	type aggRow struct {
		SlotKey   string `gorm:"column:slot_key"`
		EventType string `gorm:"column:event_type"`
		Count     int64  `gorm:"column:count"`
	}
	var agg []aggRow
	// 按广告位聚合（仅累计）
	if err := database.DB.Model(&model.AdEvent{}).
		Select("slot_key, event_type, count(*) as count").
		Group("slot_key, event_type").
		Order("slot_key").
		Scan(&agg).Error; err != nil {
		response.ServerError(c, "failed to load ad stats")
		return
	}

	type slotStat struct {
		SlotKey          string `json:"slotKey"`
		Expose           int64  `json:"expose"`
		Click            int64  `json:"click"`
		RewardedComplete int64  `json:"rewardedComplete"`
	}
	bySlotMap := map[string]*slotStat{}
	slotOrder := []string{}
	for _, r := range agg {
		st, ok := bySlotMap[r.SlotKey]
		if !ok {
			st = &slotStat{SlotKey: r.SlotKey}
			bySlotMap[r.SlotKey] = st
			slotOrder = append(slotOrder, r.SlotKey)
		}
		switch r.EventType {
		case model.AdEventExpose:
			st.Expose = r.Count
		case model.AdEventClick:
			st.Click = r.Count
		case model.AdEventRewardedComplete:
			st.RewardedComplete = r.Count
		}
	}
	bySlot := make([]slotStat, 0, len(slotOrder))
	for _, key := range slotOrder {
		bySlot = append(bySlot, *bySlotMap[key])
	}

	response.Success(c, gin.H{
		"summary": gin.H{
			// 今日曝光
			"expose": statTodayTotal{Today: exposeToday, Total: exposeTotal},
			// 今日点击
			"click": statTodayTotal{Today: clickToday, Total: clickTotal},
			// 今日激励完成
			"rewardedComplete": statTodayTotal{Today: rewardedToday, Total: rewardedTotal},
		},
		// 按广告位拆分
		"bySlot": bySlot,
	})
}

// startOfToday 返回服务器本地时区当日 00:00:00。
func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// countTodayTotal 统计指定 model 表今日与累计行数（按 created_at）。
func countTodayTotal(modelPtr any) (today, total int64) {
	database.DB.Model(modelPtr).Count(&total)
	database.DB.Model(modelPtr).Where("created_at >= ?", startOfToday()).Count(&today)
	return today, total
}

// countAdEventTodayTotal 按 event_type 统计 ad_events 今日与累计。
func countAdEventTodayTotal(eventType string) (today, total int64) {
	database.DB.Model(&model.AdEvent{}).Where("event_type = ?", eventType).Count(&total)
	database.DB.Model(&model.AdEvent{}).
		Where("event_type = ? AND created_at >= ?", eventType, startOfToday()).
		Count(&today)
	return today, total
}

// countDistinctViewUsers 统计 view_logs 去重用户数；since 非 nil 时仅统计该时间点之后。
// UV 仅含 user_id > 0 的已登录用户，匿名浏览不计入。
func countDistinctViewUsers(since *time.Time) (int64, error) {
	var n int64
	q := database.DB.Model(&model.ViewLog{}).Where("user_id > 0")
	if since != nil {
		q = q.Where("created_at >= ?", *since)
	}
	err := q.Distinct("user_id").Count(&n).Error
	return n, err
}
