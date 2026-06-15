// Day11：后台 Dashboard 统计（需 AdminAuth）。
package handler

import (
	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// StatsOverview 各核心表记录总量，供运营概览。
func (s *Server) StatsOverview(c *gin.Context) {
	var users int64
	var articles int64
	var essays int64
	var downloads int64
	var views int64
	var searches int64
	var shares int64
	var adEvents int64
	database.DB.Model(&model.User{}).Count(&users)
	database.DB.Model(&model.Article{}).Count(&articles)
	database.DB.Model(&model.EssayExample{}).Count(&essays)
	database.DB.Model(&model.DownloadLog{}).Count(&downloads)
	database.DB.Model(&model.ViewLog{}).Count(&views)
	database.DB.Model(&model.SearchLog{}).Count(&searches)
	database.DB.Model(&model.ShareLog{}).Count(&shares)
	database.DB.Model(&model.AdEvent{}).Count(&adEvents)
	response.Success(c, gin.H{
		"users":     users,
		"articles":  articles,
		"essays":    essays,
		"downloads": downloads,
		"views":     views,
		"searches":  searches,
		"shares":    shares,
		"adEvents":  adEvents,
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
		Select("keyword, count(*) as count"). // 统计关键词搜索次数
		Group("keyword").                     // 分组统计
		Order("count DESC").                  // 按搜索次数降序
		Limit(20).                            // 限制20条
		Scan(&rows).Error; err != nil {
		response.ServerError(c, "failed to list hot searches")
		return
	}
	response.Success(c, rows)
}
