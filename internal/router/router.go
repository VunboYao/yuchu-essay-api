// Package router 注册全部 HTTP 路由（唯一入口）。
// 后台：/admin/api/v1；小程序：/api/v1；健康检查：/health。
package router

import (
	"course_server/internal/config"
	"course_server/internal/handler"
	"course_server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	h := handler.NewServer(cfg)

	// Day1：健康检查
	r.GET("/health", h.Health)

	// Day3–7：后台管理 API（除 login 外均需 AdminAuth）
	admin := r.Group("/admin/api/v1")
	admin.POST("/auth/login", h.AdminLogin)
	protectedAdmin := admin.Group("")
	// 先鉴权，再记录写操作审计日志（Day3）
	protectedAdmin.Use(middleware.AdminAuth(cfg), middleware.AdminOperationLog())
	{
		protectedAdmin.GET("/stats/overview", h.StatsOverview)
		protectedAdmin.GET("/stats/hot-articles", h.StatsHotArticles)
		protectedAdmin.GET("/stats/hot-searches", h.StatsHotSearches)

		protectedAdmin.GET("/operation-logs", h.ListOperationLogs)

		protectedAdmin.GET("/taxonomy-units", h.ListTaxonomyUnits)
		protectedAdmin.POST("/taxonomy-units", h.CreateTaxonomyUnit)
		protectedAdmin.PUT("/taxonomy-units/:id", h.UpdateTaxonomyUnit)

		protectedAdmin.GET("/tags", h.ListTags)
		protectedAdmin.POST("/tags", h.CreateTag)
		protectedAdmin.PUT("/tags/:id", h.UpdateTag)

		protectedAdmin.GET("/articles", h.ListArticles)
		protectedAdmin.POST("/articles", h.CreateArticle)
		protectedAdmin.GET("/articles/:id", h.GetArticle)
		protectedAdmin.PUT("/articles/:id", h.UpdateArticle)
		protectedAdmin.PUT("/articles/:id/status", h.UpdateArticleStatus)
		protectedAdmin.GET("/articles/:id/tags", h.ListArticleTags)
		protectedAdmin.PUT("/articles/:id/tags", h.SetArticleTags)

		protectedAdmin.GET("/essay-examples", h.ListEssayExamples)
		protectedAdmin.POST("/essay-examples", h.CreateEssayExample)
		protectedAdmin.PUT("/essay-examples/:id", h.UpdateEssayExample)
		protectedAdmin.DELETE("/essay-examples/:id", h.DeleteEssayExample)
		protectedAdmin.GET("/essay-examples/:id/tags", h.ListEssayTags)
		protectedAdmin.PUT("/essay-examples/:id/tags", h.SetEssayTags)

		protectedAdmin.GET("/download-files", h.ListDownloadFiles)
		protectedAdmin.POST("/download-files", h.CreateDownloadFile)
		protectedAdmin.PUT("/download-files/:id", h.UpdateDownloadFile)
		protectedAdmin.DELETE("/download-files/:id", h.DeleteDownloadFile)

		protectedAdmin.GET("/home-blocks", h.ListHomeBlocks)
		protectedAdmin.POST("/home-blocks", h.UpsertHomeBlock)
		protectedAdmin.GET("/ad-slots", h.ListAdSlots)
		protectedAdmin.POST("/ad-slots", h.UpsertAdSlot)
		protectedAdmin.GET("/official-account-configs", h.ListOfficialAccountConfigs)
		protectedAdmin.POST("/official-account-configs", h.UpsertOfficialAccountConfig)
		protectedAdmin.POST("/upload/sts", h.UploadSTS)
	}

	// Day8–11：小程序公开 API（无需登录）
	api := r.Group("/api/v1")
	{
		api.POST("/wx/login", h.WxLogin)
		api.GET("/home", h.Home)
		api.GET("/categories", h.Categories)
		api.GET("/articles", h.PublicArticles)
		api.GET("/articles/:id", h.PublicArticleDetail)
		api.GET("/essays", h.PublicEssays)
		api.GET("/essays/:id", h.PublicEssayDetail)
		api.GET("/search", h.Search)
		api.GET("/download-files/:id/sign-url", h.SignDownloadURL)
		api.POST("/track/view", h.TrackView)
		api.POST("/track/share", h.TrackShare)
		api.POST("/track/search", h.TrackSearch)
		api.POST("/track/ad", h.TrackAd)
	}

	// Day8–10：小程序需登录 API（Bearer user JWT）
	user := api.Group("")
	user.Use(middleware.UserAuth(cfg))
	{
		user.GET("/mine", h.Mine)
		user.GET("/favorites", h.ListFavorites)
		user.POST("/favorites", h.AddFavorite)
		user.DELETE("/favorites", h.RemoveFavorite)
		user.GET("/history", h.History)
		user.GET("/downloads", h.Downloads)
	}

	return r
}
