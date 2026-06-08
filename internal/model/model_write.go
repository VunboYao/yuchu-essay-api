// 实体写操作辅助：UpdateMap、唯一性检查与按业务键查询（见 AGENTS.md §6）。
package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// jsonColumnValue 序列化 JSON 列写入值；Updates(map) 不走 struct 的 serializer tag。
func jsonColumnValue(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

// --- Tag ---

// UpdateMap 返回 GORM Updates 用的字段 map（支持零值写入）；type 已废弃不再更新。
func (t Tag) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"name":       t.Name,
		"sort_order": t.SortOrder,
		"status":     t.Status,
	}
}

// CheckTagNameDuplicate 检查 name 是否与其他记录冲突。
func CheckTagNameDuplicate(db *gorm.DB, name string, excludeID uint) error {
	q := db.Model(&Tag{}).Where("name = ?", name)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("tag with same name already exists")
	}
	return nil
}

// --- TaxonomyUnit ---

// UpdateMap 返回 GORM Updates 用的字段 map（支持零值写入）。
func (t TaxonomyUnit) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"textbook_version": t.TextbookVersion,
		"grade":            t.Grade,
		"semester":         t.Semester,
		"unit":             t.Unit,
		"unit_title":       t.UnitTitle,
		"sort_order":       t.SortOrder,
		"status":           t.Status,
	}
}

// CheckTaxonomyUnitDuplicate 检查同版本/年级/学期/单元组合是否已存在。
func CheckTaxonomyUnitDuplicate(db *gorm.DB, unit TaxonomyUnit, excludeID uint) error {
	q := db.Model(&TaxonomyUnit{}).Where(
		"textbook_version = ? AND grade = ? AND semester = ? AND unit = ?",
		unit.TextbookVersion, unit.Grade, unit.Semester, unit.Unit,
	)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("taxonomy unit with same textbook version, grade, semester and unit already exists")
	}
	return nil
}

// --- Article ---

// UpdateMap 返回 GORM Updates 用的字段 map（支持零值写入；不含 publishedAt 与计数）。
func (a Article) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"title":            a.Title,
		"slug":             a.Slug,
		"textbook_version": a.TextbookVersion,
		"grade":            a.Grade,
		"semester":         a.Semester,
		"unit":             a.Unit,
		"essay_topic":      a.EssayTopic,
		"content_type":     a.ContentType,
		"audience":         a.Audience,
		"summary":          a.Summary,
		"cover_img":        a.CoverImg,
		"body_images":      jsonColumnValue(a.BodyImages),
		"view_access_type": a.ViewAccessType,
		"status":           a.Status,
		"is_recommended":   a.IsRecommended,
		"is_hot":           a.IsHot,
		"sort_order":       a.SortOrder,
	}
}

// CheckArticleSlugDuplicate 检查 slug 是否与其他文章冲突（slug 为空时跳过）。
func CheckArticleSlugDuplicate(db *gorm.DB, slug string, excludeID uint) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil
	}
	q := db.Model(&Article{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("article slug already exists")
	}
	return nil
}

// --- EssayExample ---

// UpdateMap 返回 GORM Updates 用的字段 map（支持零值写入；不含计数）。
func (e EssayExample) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"article_id": e.ArticleID,
		"title":      e.Title,
		"grade":      e.Grade,
		"word_count": e.WordCount,
		"body":       e.Body,
		"comment":    e.Comment,
		"sort_order": e.SortOrder,
		"status":     e.Status,
	}
}

// CheckEssayArticleExists 校验关联专题文章是否存在；未关联（nil/0）时直接通过。
func CheckEssayArticleExists(db *gorm.DB, articleID *uint) error {
	if articleID == nil || *articleID == 0 {
		return nil
	}
	var count int64
	if err := db.Model(&Article{}).Where("id = ?", *articleID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("article not found")
	}
	return nil
}

// --- DownloadFile ---

// UpdateMap 返回 GORM Updates 用的字段 map（支持零值写入；不含 downloadCount）。
func (f DownloadFile) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"owner_type":  f.OwnerType,
		"owner_id":    f.OwnerID,
		"file_name":   f.FileName,
		"tags":        f.Tags,
		"access_type": f.AccessType,
		"file_type":   f.FileType,
		"file_url":    f.FileURL,
		"file_size":   f.FileSize,
		"need_login":  f.NeedLogin,
		"sort_order":  f.SortOrder,
		"status":      f.Status,
	}
}

// --- 运营配置 ---

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
