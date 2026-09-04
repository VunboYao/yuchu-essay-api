package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// 表注释前缀：写入 MySQL TABLE COMMENT，便于在库中区分使用端。
const (
	TableScopeAdmin      = "【后台】"
	TableScopeMini       = "【小程序】"
	TableScopeShared     = "【共用·后台写·小程序读】"
	TableScopeAdminWrite = "【后台写·小程序可读】"
)

// tableCommentEntries 顺序与 model_*.go、migrate.go 分组一致：共用 → 后台 → 小程序。
var tableCommentEntries = []struct {
	table   string
	comment string
}{
	// --- 共用 ---
	{table: "taxonomy_units", comment: TableScopeShared + "教材版本/年级/学期/单元分类"},
	{table: "tags", comment: TableScopeShared + "内容标签（文章/范文/下载等）"},
	{table: "articles", comment: TableScopeShared + "作文专题文章"},
	{table: "article_tags", comment: TableScopeShared + "文章与标签关联"},
	{table: "essay_examples", comment: TableScopeShared + "优秀范文"},
	{table: "essay_example_tags", comment: TableScopeShared + "范文与标签关联"},
	{table: "videos", comment: TableScopeShared + "视频号讲解视频"},
	{table: "video_tags", comment: TableScopeShared + "视频与标签关联"},
	{table: "download_files", comment: TableScopeShared + "下载资料元数据（含二期权限字段预留）"},
	{table: "ad_slots", comment: TableScopeShared + "广告位配置（后台配置、小程序展示）"},
	{table: "home_blocks", comment: TableScopeShared + "首页可配置入口（图标/标题/跳转）"},
	{table: "hot_search_keywords", comment: TableScopeShared + "搜索页运营热门词（后台配置）"},
	{table: "official_account_configs", comment: TableScopeShared + "公众号导流场景配置"},
	// --- 后台 ---
	{table: "admins", comment: TableScopeAdmin + "后台管理员账号"},
	{table: "admin_operation_logs", comment: TableScopeAdmin + "后台操作审计日志"},
	{table: "system_settings", comment: TableScopeAdminWrite + "系统 KV 配置项"},
	// --- 小程序 ---
	{table: "users", comment: TableScopeMini + "小程序微信用户"},
	{table: "download_logs", comment: TableScopeMini + "用户下载行为日志"},
	{table: "favorites", comment: TableScopeMini + "用户收藏"},
	{table: "view_logs", comment: TableScopeMini + "页面/内容浏览日志"},
	{table: "search_logs", comment: TableScopeMini + "搜索关键词日志"},
	{table: "share_logs", comment: TableScopeMini + "分享行为日志"},
	{table: "ad_events", comment: TableScopeMini + "广告曝光/点击等行为日志"},
	{table: "feedbacks", comment: TableScopeMini + "用户意见反馈"},
}

// ApplyTableComments 在 AutoMigrate 之后为各表写入 MySQL 表级 COMMENT（幂等，可重复执行）。
func ApplyTableComments(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}
	for _, e := range tableCommentEntries {
		// DDL 不支持占位符，须内联 COMMENT 并转义单引号
		escaped := strings.ReplaceAll(e.comment, "'", "''")
		sql := fmt.Sprintf("ALTER TABLE `%s` COMMENT = '%s'", e.table, escaped)
		if err := db.Exec(sql).Error; err != nil {
			return fmt.Errorf("comment table %s: %w", e.table, err)
		}
	}
	return nil
}
