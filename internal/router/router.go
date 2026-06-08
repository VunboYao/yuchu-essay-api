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
	// 创建 Gin 引擎
	r := gin.New()
	// 使用日志记录器、恢复中间件、跨域中间件
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	// 创建服务器实例
	h := handler.NewServer(cfg)

	// Day1：健康检查
	r.GET("/health", h.Health)

	// Day3–7：后台管理 API（除 login 外均需 AdminAuth）
	admin := r.Group("/admin/api/v1")
	admin.POST("/auth/login", h.AdminLogin)
	// 在 admin 下再建一个子分组, 用于需要鉴权的接口
	protectedAdmin := admin.Group("")
	// 先鉴权，再记录写操作审计日志（Day3）, 鉴权中间件和审计日志中间件
	protectedAdmin.Use(middleware.AdminAuth(cfg), middleware.AdminOperationLog())
	{
		protectedAdmin.GET("/stats/overview", h.StatsOverview)
		protectedAdmin.GET("/stats/hot-articles", h.StatsHotArticles)
		protectedAdmin.GET("/stats/hot-searches", h.StatsHotSearches)

		protectedAdmin.GET("/operation-logs", h.ListOperationLogs)

		// 教材分类单元: 查询
		protectedAdmin.GET("/taxonomy-units", h.ListTaxonomyUnits)
		// 教材分类单元: 创建
		protectedAdmin.POST("/taxonomy-units", h.CreateTaxonomyUnit)
		// 教材分类单元: 更新
		protectedAdmin.PUT("/taxonomy-units/:id", h.UpdateTaxonomyUnit)

		// 标签: 查询
		protectedAdmin.GET("/tags", h.ListTags)
		// 标签: 创建
		protectedAdmin.POST("/tags", h.CreateTag)
		// 标签: 更新
		protectedAdmin.PUT("/tags/:id", h.UpdateTag)

		// 文章: 列表查询
		protectedAdmin.GET("/articles", h.ListArticles)
		// 文章: 创建
		protectedAdmin.POST("/articles", h.CreateArticle)
		// 文章: 详情查询
		protectedAdmin.GET("/articles/:id", h.GetArticle)
		// 文章: 更新
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
