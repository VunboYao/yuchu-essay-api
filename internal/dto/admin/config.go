// 运营配置 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package admin

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"course_server/internal/dto"
	"course_server/internal/model"
)

// HomeBlockWriteRequest 首页入口创建/更新共用入参（见 prd/002 §6.1）。
type HomeBlockWriteRequest struct {
	Title     string `json:"title" binding:"required"`
	Icon      string `json:"icon" binding:"required"`
	NavType   string `json:"navType" binding:"required"`
	Path      string `json:"path" binding:"required"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

// CreateHomeBlockRequest 创建首页入口。
type CreateHomeBlockRequest = HomeBlockWriteRequest

// UpdateHomeBlockRequest 更新首页入口。
type UpdateHomeBlockRequest = HomeBlockWriteRequest

func (r HomeBlockWriteRequest) normalize() (model.HomeBlock, error) {
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return model.HomeBlock{}, fmt.Errorf("title is required")
	}
	if len(title) > 120 {
		return model.HomeBlock{}, fmt.Errorf("title exceeds max 120")
	}
	icon := strings.TrimSpace(r.Icon)
	if icon == "" {
		return model.HomeBlock{}, fmt.Errorf("icon is required")
	}
	if len(icon) > 500 {
		return model.HomeBlock{}, fmt.Errorf("icon exceeds max 500")
	}
	navType := strings.TrimSpace(r.NavType)
	if navType == "" {
		return model.HomeBlock{}, fmt.Errorf("navType is required")
	}
	if !model.IsValidNavType(navType) {
		return model.HomeBlock{}, fmt.Errorf("navType must be one of: navigateTo, switchTab, redirectTo, reLaunch")
	}
	path := strings.TrimSpace(r.Path)
	if path == "" {
		return model.HomeBlock{}, fmt.Errorf("path is required")
	}
	if len(path) > 500 {
		return model.HomeBlock{}, fmt.Errorf("path exceeds max 500")
	}
	if !strings.HasPrefix(path, "/") {
		return model.HomeBlock{}, fmt.Errorf("path must start with /")
	}
	return model.HomeBlock{
		Title:     title,
		Icon:      icon,
		NavType:   navType,
		Path:      path,
		SortOrder: r.SortOrder,
	}, nil
}

// ValidateForCreate 校验创建入参；status 未传时默认启用。
func (r HomeBlockWriteRequest) ValidateForCreate() (model.HomeBlock, error) {
	block, err := r.normalize()
	if err != nil {
		return model.HomeBlock{}, err
	}
	status, err := dto.ResolveStatus(r.Status, model.StatusEnabled)
	if err != nil {
		return model.HomeBlock{}, err
	}
	block.Status = status
	return block, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r HomeBlockWriteRequest) ValidateForUpdate(existing model.HomeBlock) (model.HomeBlock, error) {
	block, err := r.normalize()
	if err != nil {
		return model.HomeBlock{}, err
	}
	status, err := dto.ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return model.HomeBlock{}, err
	}
	block.Status = status
	return block, nil
}

const maxHotSearchKeywordRunes = 40

// HotSearchKeywordWriteRequest 搜索热门词创建/更新共用入参。
type HotSearchKeywordWriteRequest struct {
	Keyword   string `json:"keyword" binding:"required"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

// CreateHotSearchKeywordRequest 创建搜索热门词。
type CreateHotSearchKeywordRequest = HotSearchKeywordWriteRequest

// UpdateHotSearchKeywordRequest 更新搜索热门词。
type UpdateHotSearchKeywordRequest = HotSearchKeywordWriteRequest

func (r HotSearchKeywordWriteRequest) normalize() (model.HotSearchKeyword, error) {
	keyword := strings.TrimSpace(r.Keyword)
	if keyword == "" {
		return model.HotSearchKeyword{}, fmt.Errorf("keyword is required")
	}
	if utf8.RuneCountInString(keyword) > maxHotSearchKeywordRunes {
		return model.HotSearchKeyword{}, fmt.Errorf("keyword exceeds max %d", maxHotSearchKeywordRunes)
	}
	return model.HotSearchKeyword{
		Keyword:   keyword,
		SortOrder: r.SortOrder,
	}, nil
}

// ValidateForCreate 校验创建入参；status 未传时默认启用。
func (r HotSearchKeywordWriteRequest) ValidateForCreate() (model.HotSearchKeyword, error) {
	item, err := r.normalize()
	if err != nil {
		return model.HotSearchKeyword{}, err
	}
	status, err := dto.ResolveStatus(r.Status, model.StatusEnabled)
	if err != nil {
		return model.HotSearchKeyword{}, err
	}
	item.Status = status
	return item, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r HotSearchKeywordWriteRequest) ValidateForUpdate(existing model.HotSearchKeyword) (model.HotSearchKeyword, error) {
	item, err := r.normalize()
	if err != nil {
		return model.HotSearchKeyword{}, err
	}
	status, err := dto.ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return model.HotSearchKeyword{}, err
	}
	item.Status = status
	return item, nil
}

// UpsertAdSlotRequest 广告位 upsert 入参（按 slotKey 匹配已有记录）。
type UpsertAdSlotRequest struct {
	SlotKey   string `json:"slotKey" binding:"required"`
	SlotName  string `json:"slotName" binding:"required"`
	Page      string `json:"page" binding:"required"`
	Position  string `json:"position"`
	AdType    string `json:"adType"`
	AdUnitID  string `json:"adUnitId"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

func (r UpsertAdSlotRequest) normalize() (model.AdSlot, error) {
	slotKey := strings.TrimSpace(r.SlotKey)
	if slotKey == "" {
		return model.AdSlot{}, fmt.Errorf("slotKey is required")
	}
	slotName := strings.TrimSpace(r.SlotName)
	if slotName == "" {
		return model.AdSlot{}, fmt.Errorf("slotName is required")
	}
	page := strings.TrimSpace(r.Page)
	if page == "" {
		return model.AdSlot{}, fmt.Errorf("page is required")
	}
	adType := strings.TrimSpace(r.AdType)
	if adType == "" {
		adType = model.AdTypeBanner
	}
	if !model.ValidAdType(adType) {
		return model.AdSlot{}, fmt.Errorf("adType must be banner or rewarded_video")
	}
	if adType == model.AdTypeRewardedVideo && page != model.AdPageDownload {
		return model.AdSlot{}, fmt.Errorf("rewarded_video ad slot page must be download")
	}
	return model.AdSlot{
		SlotKey:   slotKey,
		SlotName:  slotName,
		Page:      page,
		Position:  strings.TrimSpace(r.Position),
		AdType:    adType,
		AdUnitID:  strings.TrimSpace(r.AdUnitID),
		SortOrder: r.SortOrder,
	}, nil
}

func (r UpsertAdSlotRequest) Validate(defaultStatus int) (model.AdSlot, error) {
	slot, err := r.normalize()
	if err != nil {
		return model.AdSlot{}, err
	}
	status, err := dto.ResolveStatus(r.Status, defaultStatus)
	if err != nil {
		return model.AdSlot{}, err
	}
	slot.Status = status
	return slot, nil
}

// UpsertOfficialAccountConfigRequest 公众号导流 upsert 入参（按 scene 匹配已有记录）。
type UpsertOfficialAccountConfigRequest struct {
	Scene       string `json:"scene" binding:"required"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Username    string `json:"username" binding:"required"`
	Status      *int   `json:"status"`
}

func (r UpsertOfficialAccountConfigRequest) normalize() (model.OfficialAccountConfig, error) {
	scene := strings.TrimSpace(r.Scene)
	if scene == "" {
		return model.OfficialAccountConfig{}, fmt.Errorf("scene is required")
	}
	username := strings.TrimSpace(r.Username)
	if username == "" {
		return model.OfficialAccountConfig{}, fmt.Errorf("username is required")
	}
	return model.OfficialAccountConfig{
		Scene:       scene,
		Title:       strings.TrimSpace(r.Title),
		Description: strings.TrimSpace(r.Description),
		Username:    username,
	}, nil
}

func (r UpsertOfficialAccountConfigRequest) Validate(defaultStatus int) (model.OfficialAccountConfig, error) {
	cfg, err := r.normalize()
	if err != nil {
		return model.OfficialAccountConfig{}, err
	}
	status, err := dto.ResolveStatus(r.Status, defaultStatus)
	if err != nil {
		return model.OfficialAccountConfig{}, err
	}
	cfg.Status = status
	return cfg, nil
}
