// 实体写操作辅助：UpdateMap、唯一性检查与按业务键查询（见 AGENTS.md §6）。
package model

import (
	"encoding/json"
	"errors"
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

// CountTagUsage 统计标签被文章与优秀作文引用的次数。
func CountTagUsage(db *gorm.DB, tagID uint) (articleCount, essayCount int64, err error) {
	if err = db.Model(&ArticleTag{}).Where("tag_id = ?", tagID).Count(&articleCount).Error; err != nil {
		return
	}
	err = db.Model(&EssayExampleTag{}).Where("tag_id = ?", tagID).Count(&essayCount).Error
	return
}

// tagInUseError 标签仍被内容引用时的中文错误提示。
func tagInUseError(articleCount, essayCount int64) error {
	if articleCount == 0 && essayCount == 0 {
		return nil
	}
	return fmt.Errorf("该标签仍被 %d 篇文章、%d 篇优秀作文使用，请先在对应内容中取消关联后再操作", articleCount, essayCount)
}

// CheckTagDisableAllowed 启用→禁用前校验：无内容关联才允许禁用。
func CheckTagDisableAllowed(db *gorm.DB, tagID uint) error {
	articleCount, essayCount, err := CountTagUsage(db, tagID)
	if err != nil {
		return err
	}
	return tagInUseError(articleCount, essayCount)
}

// CheckTagDeletable 物理删除前校验：须已禁用且无内容关联。
func CheckTagDeletable(db *gorm.DB, tag Tag) error {
	if tag.Status != StatusDisabled {
		return fmt.Errorf("仅已禁用的标签可删除，请先将标签禁用")
	}
	return CheckTagDisableAllowed(db, tag.ID)
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

// CheckDownloadFileDeletable 校验下载资料是否可物理删除。
// 已有下载记录时默认拒绝删除，建议 PUT 禁用或去掉「免费」标签；传 force 时可强制删除。
func CheckDownloadFileDeletable(db *gorm.DB, file DownloadFile, force bool) error {
	if force {
		return nil
	}
	if file.DownloadCount > 0 {
		return fmt.Errorf("download file has download history, set status to disabled or pass confirm=true")
	}
	var logCount int64
	if err := db.Model(&DownloadLog{}).Where("file_id = ?", file.ID).Count(&logCount).Error; err != nil {
		return err
	}
	if logCount > 0 {
		return fmt.Errorf("download file has download history, set status to disabled or pass confirm=true")
	}
	return nil
}

// CheckDownloadOwnerExists 校验资料归属资源是否存在；一期仅支持 owner_type=article。
func CheckDownloadOwnerExists(db *gorm.DB, ownerType string, ownerID uint) error {
	switch ownerType {
	case TargetTypeArticle:
		var count int64
		if err := db.Model(&Article{}).Where("id = ?", ownerID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("article not found")
		}
		return nil
	default:
		return fmt.Errorf("invalid ownerType")
	}
}

var (
	// ErrInvalidFavoriteTargetType 收藏目标类型非法（非 article/essay）。
	ErrInvalidFavoriteTargetType = errors.New("invalid targetType")
	// ErrFavoriteArticleNotFound 文章不存在或未发布，不可收藏。
	ErrFavoriteArticleNotFound = errors.New("article not found")
	// ErrFavoriteEssayNotFound 优秀作文不存在或未启用，不可收藏。
	ErrFavoriteEssayNotFound = errors.New("essay not found")
	// ErrFavoriteNotFound 当前用户无对应收藏记录。
	ErrFavoriteNotFound = errors.New("favorite not found")
)

// ValidateFavoriteTargetType 校验收藏目标类型是否为 article 或 essay。
func ValidateFavoriteTargetType(targetType string) error {
	switch targetType {
	case TargetTypeArticle, TargetTypeEssay:
		return nil
	default:
		return ErrInvalidFavoriteTargetType
	}
}

// CheckFavoriteTargetExists 校验 article/essay 目标存在且对小程序可见；收藏与埋点共用。
func CheckFavoriteTargetExists(db *gorm.DB, targetType string, targetID uint) error {
	if err := ValidateFavoriteTargetType(targetType); err != nil {
		return err
	}
	switch targetType {
	case TargetTypeArticle:
		var count int64
		if err := db.Model(&Article{}).
			Where("id = ? AND status = ?", targetID, ArticleStatusPublished).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrFavoriteArticleNotFound
		}
		return nil
	case TargetTypeEssay:
		var count int64
		if err := db.Model(&EssayExample{}).
			Where("id = ? AND status = ?", targetID, StatusEnabled).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrFavoriteEssayNotFound
		}
		return nil
	default:
		return ErrInvalidFavoriteTargetType
	}
}

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
		"title":      b.Title,
		"icon":       b.Icon,
		"nav_type":   b.NavType,
		"path":       b.Path,
		"sort_order": b.SortOrder,
		"status":     b.Status,
	}
}

// FindHomeBlockByID 按 id 查找首页入口；不存在时返回 gorm.ErrRecordNotFound。
func FindHomeBlockByID(db *gorm.DB, id uint) (HomeBlock, error) {
	var block HomeBlock
	err := db.First(&block, id).Error
	return block, err
}

func (s AdSlot) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"slot_key":   s.SlotKey,
		"slot_name":  s.SlotName,
		"page":       s.Page,
		"position":   s.Position,
		"ad_type":    s.AdType,
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

// FindDownloadGateSlot 查询资料下载激励视频门槛配置（固定 slotKey）。
func FindDownloadGateSlot(db *gorm.DB) (AdSlot, error) {
	return FindAdSlotByKey(db, AdSlotKeyDownloadRewardedVideo)
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
