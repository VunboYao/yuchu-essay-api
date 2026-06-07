// 运营配置 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package admin

import (
	"fmt"
	"strings"

	"course_server/internal/dto"
	"course_server/internal/model"
)

// UpsertHomeBlockRequest 首页区块 upsert 入参（按 blockKey 匹配已有记录）。
type UpsertHomeBlockRequest struct {
	BlockKey    string `json:"blockKey" binding:"required"`
	Title       string `json:"title" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
	ConfigJSON  string `json:"configJson"`
	SortOrder   int    `json:"sortOrder"`
	Status      *int   `json:"status"`
}

func (r UpsertHomeBlockRequest) normalize() (model.HomeBlock, error) {
	blockKey := strings.TrimSpace(r.BlockKey)
	if blockKey == "" {
		return model.HomeBlock{}, fmt.Errorf("blockKey is required")
	}
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return model.HomeBlock{}, fmt.Errorf("title is required")
	}
	contentType := strings.TrimSpace(r.ContentType)
	if contentType == "" {
		return model.HomeBlock{}, fmt.Errorf("contentType is required")
	}
	return model.HomeBlock{
		BlockKey:    blockKey,
		Title:       title,
		ContentType: contentType,
		ConfigJSON:  r.ConfigJSON,
		SortOrder:   r.SortOrder,
	}, nil
}

func (r UpsertHomeBlockRequest) Validate(defaultStatus int) (model.HomeBlock, error) {
	block, err := r.normalize()
	if err != nil {
		return model.HomeBlock{}, err
	}
	status, err := dto.ResolveStatus(r.Status, defaultStatus)
	if err != nil {
		return model.HomeBlock{}, err
	}
	block.Status = status
	return block, nil
}

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
	return model.AdSlot{
		SlotKey:   slotKey,
		SlotName:  slotName,
		Page:      page,
		Position:  strings.TrimSpace(r.Position),
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
	QrcodeURL   string `json:"qrcodeUrl"`
	LinkText    string `json:"linkText"`
	Status      *int   `json:"status"`
}

func (r UpsertOfficialAccountConfigRequest) normalize() (model.OfficialAccountConfig, error) {
	scene := strings.TrimSpace(r.Scene)
	if scene == "" {
		return model.OfficialAccountConfig{}, fmt.Errorf("scene is required")
	}
	return model.OfficialAccountConfig{
		Scene:       scene,
		Title:       strings.TrimSpace(r.Title),
		Description: strings.TrimSpace(r.Description),
		QrcodeURL:   strings.TrimSpace(r.QrcodeURL),
		LinkText:    strings.TrimSpace(r.LinkText),
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
