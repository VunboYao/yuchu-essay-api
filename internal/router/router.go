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
		// 统计概览
		protectedAdmin.GET("/stats/overview", h.StatsOverview)
		// 热门文章10条
		protectedAdmin.GET("/stats/hot-articles", h.StatsHotArticles)
		// 热门搜索词
		protectedAdmin.GET("/stats/hot-searches", h.StatsHotSearches)
		// 免费资料下载排行
		protectedAdmin.GET("/stats/hot-downloads", h.StatsHotDownloads)
		// 广告曝光/点击统计
		protectedAdmin.GET("/stats/ad-events", h.StatsAdEvents)
		// 操作日志
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
		// 文章: 更新状态
		protectedAdmin.PUT("/articles/:id/status", h.UpdateArticleStatus)

		// 优秀作文: 列表查询
		protectedAdmin.GET("/essay-examples", h.ListEssayExamples)
		// 优秀作文: 创建
		protectedAdmin.POST("/essay-examples", h.CreateEssayExample)
		// 优秀作文: 详情查询
		protectedAdmin.GET("/essay-examples/:id", h.GetEssayExample)
		// 优秀作文: 更新
		protectedAdmin.PUT("/essay-examples/:id", h.UpdateEssayExample)
		// 优秀作文: 删除
		protectedAdmin.DELETE("/essay-examples/:id", h.DeleteEssayExample)

		// 下载文件: 列表查询
		protectedAdmin.GET("/download-files", h.ListDownloadFiles)
		// 下载文件: 创建
		protectedAdmin.POST("/download-files", h.CreateDownloadFile)
		// 下载文件: 更新
		protectedAdmin.PUT("/download-files/:id", h.UpdateDownloadFile)
		// 文件: 删除
		protectedAdmin.DELETE("/download-files/:id", h.DeleteDownloadFile)

		// 首页入口: 列表查询
		protectedAdmin.GET("/home-blocks", h.ListHomeBlocks)
		// 首页入口: 创建
		protectedAdmin.POST("/home-blocks", h.CreateHomeBlock)
		// 首页入口: 更新
		protectedAdmin.PUT("/home-blocks/:id", h.UpdateHomeBlock)
		// 广告位: 列表查询
		protectedAdmin.GET("/ad-slots", h.ListAdSlots)
		// 广告位: 创建或更新
		protectedAdmin.POST("/ad-slots", h.UpsertAdSlot)
		// 公众号配置: 列表查询
		protectedAdmin.GET("/official-account-configs", h.ListOfficialAccountConfigs)
		// 公众号配置: 创建或更新
		protectedAdmin.POST("/official-account-configs", h.UpsertOfficialAccountConfig)
		// 上传STS凭证
		protectedAdmin.POST("/upload/sts", h.UploadSTS)
	}

	// Day8–11：小程序公开 API（无需登录）
	api := r.Group("/api/v1")
	{
		// 小程序登录
		api.POST("/wx/login", h.WxLogin)
		// 首页
		api.GET("/home", h.Home)
		// 分类
		api.GET("/categories", h.Categories)
		// 查询已发布的文章
		api.GET("/articles", h.PublicArticles)
		// 查询已发布文章详情
		api.GET("/articles/:id", h.PublicArticleDetail)
		// 查询优秀作文
		api.GET("/essays", h.PublicEssays)
		// 查询优秀作文详情
		api.GET("/essays/:id", h.PublicEssayDetail)
		// 搜索
		api.GET("/search", h.Search)
		// 广告位配置（仅返回启用项，可选 ?page= 过滤）
		api.GET("/ad-slots", h.PublicAdSlots)

		// 下载：可选 JWT，download_logs / ad_events 的 userId 仅服务端从 token 解析
		download := api.Group("")
		download.Use(middleware.OptionalUserAuth(cfg))
		{
			download.GET("/download-files/:id/sign-url", h.SignDownloadURL)
			download.POST("/download-files/:id/ad-unlock", h.UnlockDownloadAd)
		}

		// 埋点：可选 JWT，userId 仅服务端从 token 解析（未登录为 0）
		track := api.Group("")
		track.Use(middleware.OptionalUserAuth(cfg))
		{
			// 浏览
			track.POST("/track/view", h.TrackView)
			// 分享
			track.POST("/track/share", h.TrackShare)
			// 搜索: 在独立搜索接口中实现，这里只记录日志
			track.POST("/track/search", h.TrackSearch)
			// 广告
			track.POST("/track/ad", h.TrackAd)
		}
	}

	// Day8–10：小程序需登录 API（Bearer user JWT）
	user := api.Group("")
	user.Use(middleware.UserAuth(cfg))
	{
		// 我的页概览
		user.GET("/mine", h.Mine)
		// 我的收藏
		user.GET("/favorites", h.ListFavorites)
		// 收藏文章或优秀作文
		user.POST("/favorites", h.AddFavorite)
		// 取消收藏文章或优秀作文
		user.DELETE("/favorites", h.RemoveFavorite)
		// 浏览历史
		user.GET("/history", h.History)
		// 下载记录
		user.GET("/downloads", h.Downloads)
	}

	return r
}
