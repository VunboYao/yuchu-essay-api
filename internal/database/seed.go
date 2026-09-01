package database

import (
	"errors"

	"course_server/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// SeedAdmin 初始化默认管理员（Day2/Day3）。
// 管理员：admin/admin123；标签需通过 Admin API 创建。
// 注意：taxonomy_units 未在此种子，需后台或后续脚本维护。
func SeedAdmin() error {
	if DB == nil {
		return errors.New("database is not initialized")
	}

	var count int64
	// 检查是否已经存在管理员
	if err := DB.Model(&model.Admin{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return err
	}
	// 如果管理员不存在，则创建管理员
	if count == 0 {
		// 生成密码哈希
		hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		// 创建管理员
		if err := DB.Create(&model.Admin{
			Username:     "admin",
			PasswordHash: string(hash),
			Role:         "admin",
			Status:       model.StatusEnabled,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

// SeedAdSlots 初始化 PRD 一期广告位（幂等：仅补缺失的 slotKey）。
// 默认禁用且 adUnitId 为空，运营在微信流量主后台创建单元后通过 Admin API 填写并启用。
func SeedAdSlots() error {
	if DB == nil {
		return errors.New("database is not initialized")
	}

	defaults := []model.AdSlot{
		{
			SlotKey:   model.AdSlotKeyHomeFeedMid,
			SlotName:  "首页信息流中部",
			Page:      model.AdPageHome,
			Position:  "feed_mid",
			AdType:    model.AdTypeBanner,
			Status:    model.StatusDisabled,
			SortOrder: 40,
		},
		{
			SlotKey:   model.AdSlotKeyArticleDetailMid,
			SlotName:  "作文专题详情正文中后段",
			Page:      model.AdPageArticleDetail,
			Position:  "body_mid",
			AdType:    model.AdTypeBanner,
			Status:    model.StatusDisabled,
			SortOrder: 30,
		},
		{
			SlotKey:   model.AdSlotKeyEssayDetailBottom,
			SlotName:  "优秀作文详情底部",
			Page:      model.AdPageEssayDetail,
			Position:  "bottom",
			AdType:    model.AdTypeBanner,
			Status:    model.StatusDisabled,
			SortOrder: 20,
		},
		{
			SlotKey:   model.AdSlotKeySearchResultMid,
			SlotName:  "搜索结果页中部",
			Page:      model.AdPageSearch,
			Position:  "result_mid",
			AdType:    model.AdTypeBanner,
			Status:    model.StatusDisabled,
			SortOrder: 10,
		},
		{
			SlotKey:   model.AdSlotKeyMineTop,
			SlotName:  "我的页顶部",
			Page:      model.AdPageMine,
			Position:  "top",
			AdType:    model.AdTypeBanner,
			Status:    model.StatusDisabled,
			SortOrder: 15,
		},
		{
			SlotKey:   model.AdSlotKeyDownloadRewardedVideo,
			SlotName:  "资料下载激励视频",
			Page:      model.AdPageDownload,
			Position:  "gate",
			AdType:    model.AdTypeRewardedVideo,
			Status:    model.StatusDisabled,
			SortOrder: 50,
		},
	}

	for _, slot := range defaults {
		var count int64
		if err := DB.Model(&model.AdSlot{}).Where("slot_key = ?", slot.SlotKey).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			// 显式 Select，避免 status=0 被 GORM 当作零值省略后落库为 default:1（启用）
			if err := DB.Select(
				"SlotKey", "SlotName", "Page", "Position", "AdType", "Status", "SortOrder",
			).Create(&slot).Error; err != nil {
				return err
			}
			continue
		}
		// 预置位未配置 adUnitId 时不应保持启用（修正历史 GORM 零值或误操作）
		if err := DB.Model(&model.AdSlot{}).
			Where(
				"slot_key = ? AND status = ? AND (ad_unit_id = '' OR ad_unit_id IS NULL)",
				slot.SlotKey, model.StatusEnabled,
			).
			Update("status", model.StatusDisabled).Error; err != nil {
			return err
		}
	}
	return nil
}
