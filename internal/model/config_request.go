// 运营配置 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// --- HomeBlock ---

// UpsertHomeBlockRequest 首页区块 upsert 入参（按 blockKey 匹配已有记录）。
type UpsertHomeBlockRequest struct {
	BlockKey    string `json:"blockKey" binding:"required"`
	Title       string `json:"title" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
	ConfigJSON  string `json:"configJson"`
	SortOrder   int    `json:"sortOrder"`
	Status      *int   `json:"status"`
}

func (r UpsertHomeBlockRequest) normalize() (HomeBlock, error) {
	blockKey := strings.TrimSpace(r.BlockKey)
	if blockKey == "" {
		return HomeBlock{}, fmt.Errorf("blockKey is required")
	}
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return HomeBlock{}, fmt.Errorf("title is required")
	}
	contentType := strings.TrimSpace(r.ContentType)
	if contentType == "" {
		return HomeBlock{}, fmt.Errorf("contentType is required")
	}
	return HomeBlock{
		BlockKey:    blockKey,
		Title:       title,
		ContentType: contentType,
		ConfigJSON:  r.ConfigJSON,
		SortOrder:   r.SortOrder,
	}, nil
}

func (r UpsertHomeBlockRequest) Validate(defaultStatus int) (HomeBlock, error) {
	block, err := r.normalize()
	if err != nil {
		return HomeBlock{}, err
	}
	status, err := ResolveStatus(r.Status, defaultStatus)
	if err != nil {
		return HomeBlock{}, err
	}
	block.Status = status
	return block, nil
}

func (b HomeBlock) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"block_key":    b.BlockKey,
		"title":        b.Title,
		"content_type": b.ContentType,
		"config_json":  b.ConfigJSON,
		"sort_order":   b.SortOrder,
		"status":       b.Status,
	}
}

// FindHomeBlockByKey 按 blockKey 查找区块；不存在时返回 gorm.ErrRecordNotFound。
func FindHomeBlockByKey(db *gorm.DB, blockKey string) (HomeBlock, error) {
	var block HomeBlock
	err := db.Where("block_key = ?", blockKey).First(&block).Error
	return block, err
}

// --- AdSlot ---

// UpsertAdSlotRequest 广告位 upsert 入参（按 slotKey 匹配已有记录）。
type UpsertAdSlotRequest struct {
	SlotKey   string `json:"slotKey" binding:"required"`
	SlotName  string `json:"slotName" binding:"required"`
	Page      string `json:"page" binding:"required"`
	Position  string `json:"position"`
	AdUnitID  string `json:"adUnitId"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

func (r UpsertAdSlotRequest) normalize() (AdSlot, error) {
	slotKey := strings.TrimSpace(r.SlotKey)
	if slotKey == "" {
		return AdSlot{}, fmt.Errorf("slotKey is required")
	}
	slotName := strings.TrimSpace(r.SlotName)
	if slotName == "" {
		return AdSlot{}, fmt.Errorf("slotName is required")
	}
	page := strings.TrimSpace(r.Page)
	if page == "" {
		return AdSlot{}, fmt.Errorf("page is required")
	}
	return AdSlot{
		SlotKey:   slotKey,
		SlotName:  slotName,
		Page:      page,
		Position:  strings.TrimSpace(r.Position),
		AdUnitID:  strings.TrimSpace(r.AdUnitID),
		SortOrder: r.SortOrder,
	}, nil
}

func (r UpsertAdSlotRequest) Validate(defaultStatus int) (AdSlot, error) {
	slot, err := r.normalize()
	if err != nil {
		return AdSlot{}, err
	}
	status, err := ResolveStatus(r.Status, defaultStatus)
	if err != nil {
		return AdSlot{}, err
	}
	slot.Status = status
	return slot, nil
}

func (s AdSlot) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"slot_key":   s.SlotKey,
		"slot_name":  s.SlotName,
		"page":       s.Page,
		"position":   s.Position,
		"ad_unit_id": s.AdUnitID,
		"sort_order": s.SortOrder,
		"status":     s.Status,
	}
}

func FindAdSlotByKey(db *gorm.DB, slotKey string) (AdSlot, error) {
	var slot AdSlot
	err := db.Where("slot_key = ?", slotKey).First(&slot).Error
	return slot, err
}

// --- OfficialAccountConfig ---

// UpsertOfficialAccountConfigRequest 公众号导流 upsert 入参（按 scene 匹配已有记录）。
type UpsertOfficialAccountConfigRequest struct {
	Scene       string `json:"scene" binding:"required"`
	Title       string `json:"title"`
	Description string `json:"description"`
	QrcodeURL   string `json:"qrcodeUrl"`
	LinkText    string `json:"linkText"`
	Status      *int   `json:"status"`
}

func (r UpsertOfficialAccountConfigRequest) normalize() (OfficialAccountConfig, error) {
	scene := strings.TrimSpace(r.Scene)
	if scene == "" {
		return OfficialAccountConfig{}, fmt.Errorf("scene is required")
	}
	return OfficialAccountConfig{
		Scene:       scene,
		Title:       strings.TrimSpace(r.Title),
		Description: strings.TrimSpace(r.Description),
		QrcodeURL:   strings.TrimSpace(r.QrcodeURL),
		LinkText:    strings.TrimSpace(r.LinkText),
	}, nil
}

func (r UpsertOfficialAccountConfigRequest) Validate(defaultStatus int) (OfficialAccountConfig, error) {
	cfg, err := r.normalize()
	if err != nil {
		return OfficialAccountConfig{}, err
	}
	status, err := ResolveStatus(r.Status, defaultStatus)
	if err != nil {
		return OfficialAccountConfig{}, err
	}
	cfg.Status = status
	return cfg, nil
}

func (c OfficialAccountConfig) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"scene":       c.Scene,
		"title":       c.Title,
		"description": c.Description,
		"qrcode_url":  c.QrcodeURL,
		"link_text":   c.LinkText,
		"status":      c.Status,
	}
}

func FindOfficialAccountConfigByScene(db *gorm.DB, scene string) (OfficialAccountConfig, error) {
	var cfg OfficialAccountConfig
	err := db.Where("scene = ?", scene).First(&cfg).Error
	return cfg, err
}
