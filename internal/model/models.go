// Package model 定义一期 GORM 模型与业务常量，按使用端分文件：
//
//	model_shared.go      — 【共用】内容/配置，后台维护、小程序读取
//	model_admin.go       — 【后台】管理员与审计
//	model_miniprogram.go — 【小程序】用户与行为日志
//	model_order.go       — 【二期】订单/支付（占位，一期勿实现）
//
// 字段 gorm:"comment:..." 同步列注释；表级 COMMENT 见 db_comments.go。
package model

// 通用与内容状态、下载权限常量（禁止在 handler 中散落魔法值）。
const (
	// 通用状态常量: 0禁用 1启用
	StatusDisabled = 0
	StatusEnabled  = 1

	// 文章状态常量：0草稿 1已发布
	ArticleStatusDraft     = 0
	ArticleStatusPublished = 1

	// 下载分类常量：免费资料 课件资料 会员资料 内部资料
	DownloadCategoryFree       = "free"
	DownloadCategoryCourseware = "courseware"
	DownloadCategoryMember     = "member"
	DownloadCategoryInternal   = "internal"

	// 下载权限常量：免费 付费 会员 内部
	AccessTypeFree             = "free"
	AccessTypePurchaseRequired = "purchase_required"
	AccessTypeMemberOnly       = "member_only"
	AccessTypeInternal         = "internal" // 一期一律拒绝公开下载签名

	// 收藏/分享/浏览等行为的目标类型：文章 作文
	TargetTypeArticle = "article"
	TargetTypeEssay   = "essay"

	// 标签类型：article 文章 essay 范文 download 下载 topic 主题
	TagTypeArticle  = "article"
	TagTypeEssay    = "essay"
	TagTypeDownload = "download"
	TagTypeTopic    = "topic"

	// 最大正文图片数量
	MaxBodyImages = 20

	// 默认教材版本（统编版）
	DefaultTextbookVersion = "统编版"
)
