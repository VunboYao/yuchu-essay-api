// Package model 定义一期 GORM 实体、业务常量与持久化辅助，按使用端分文件：
//
//	model_shared.go      — 【共用】内容/配置，后台维护、小程序读取
//	model_admin.go       — 【后台】管理员与审计
//	model_miniprogram.go — 【小程序】用户与行为日志
//	model_order.go       — 【二期】订单/支付（占位，一期勿实现）
//	model_write.go       — UpdateMap、唯一性检查、按业务键查询
//
// HTTP 入参 DTO 见 internal/dto/admin/；字段 gorm:"comment:..." 同步列注释；表级 COMMENT 见 db_comments.go。
package model

// 通用与内容状态、下载权限常量（禁止在 handler 中散落魔法值）。
const (
	// 通用状态常量: 0禁用 1启用
	StatusDisabled = 0
	StatusEnabled  = 1

	// 文章状态常量：0草稿 1已发布
	ArticleStatusDraft     = 0
	ArticleStatusPublished = 1

	// 下载权限常量（资料分类与访问权限统一由此字段表达）：免费 付费 会员 内部
	AccessTypeFree             = "free"
	AccessTypePurchaseRequired = "purchase_required"
	AccessTypeMemberOnly       = "member_only"
	AccessTypeInternal         = "internal" // 一期一律拒绝公开下载签名

	// 文章查看权限常量（与下载权限是两条独立权限线）：
	// 一期固定 free（全部免费查看），member_only/purchase_required 为二期预留容错。
	ViewAccessFree             = "free"
	ViewAccessMemberOnly       = "member_only"
	ViewAccessPurchaseRequired = "purchase_required"

	// 收藏/分享/浏览等行为的目标类型：文章 作文 下载资料
	TargetTypeArticle      = "article"
	TargetTypeEssay         = "essay"
	TargetTypeDownloadFile  = "download_file"

	// 最大正文图片数量
	MaxBodyImages = 20

	// 默认教材版本（统编版）
	DefaultTextbookVersion = "统编版"

	// 学期册次：上册 / 下册
	SemesterFirst  = "上册"
	SemesterSecond = "下册"
)

// ValidSemester 校验学期是否为 PRD 定义的上册/下册之一。
func ValidSemester(s string) bool {
	switch s {
	case SemesterFirst, SemesterSecond:
		return true
	default:
		return false
	}
}

// ValidViewAccessType 校验文章查看权限是否为预留的三种之一（一期仅用 free）。
func ValidViewAccessType(t string) bool {
	switch t {
	case ViewAccessFree, ViewAccessMemberOnly, ViewAccessPurchaseRequired:
		return true
	default:
		return false
	}
}
