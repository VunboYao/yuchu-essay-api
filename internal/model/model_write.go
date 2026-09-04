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

// CountTagUsage 统计标签被文章、优秀作文与讲解视频引用的次数。
func CountTagUsage(db *gorm.DB, tagID uint) (articleCount, essayCount, videoCount int64, err error) {
	if err = db.Model(&ArticleTag{}).Where("tag_id = ?", tagID).Count(&articleCount).Error; err != nil {
		return
	}
	if err = db.Model(&EssayExampleTag{}).Where("tag_id = ?", tagID).Count(&essayCount).Error; err != nil {
		return
	}
	err = db.Model(&VideoTag{}).Where("tag_id = ?", tagID).Count(&videoCount).Error
	return
}

// tagInUseError 标签仍被内容引用时的中文错误提示。
func tagInUseError(articleCount, essayCount, videoCount int64) error {
	if articleCount == 0 && essayCount == 0 && videoCount == 0 {
		return nil
	}
	return fmt.Errorf("该标签仍被 %d 篇文章、%d 篇优秀作文、%d 条讲解视频使用，请先在对应内容中取消关联后再操作", articleCount, essayCount, videoCount)
}

// CheckTagDisableAllowed 启用→禁用前校验：无内容关联才允许禁用。
func CheckTagDisableAllowed(db *gorm.DB, tagID uint) error {
	articleCount, essayCount, videoCount, err := CountTagUsage(db, tagID)
	if err != nil {
		return err
	}
	return tagInUseError(articleCount, essayCount, videoCount)
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

// CountTaxonomyUnitArticleUsage 统计占用同一教材版本/年级/学期/单元的文章数。
func CountTaxonomyUnitArticleUsage(db *gorm.DB, unit TaxonomyUnit) (int64, error) {
	var count int64
	err := db.Model(&Article{}).Where(
		"textbook_version = ? AND grade = ? AND semester = ? AND unit = ?",
		unit.TextbookVersion, unit.Grade, unit.Semester, unit.Unit,
	).Count(&count).Error
	return count, err
}

// CheckTaxonomyUnitDeletable 物理删除前校验：须已禁用且无文章占用同一四元组。
func CheckTaxonomyUnitDeletable(db *gorm.DB, unit TaxonomyUnit) error {
	if unit.Status != StatusDisabled {
		return fmt.Errorf("仅已禁用的分类单元可删除，请先将分类单元禁用")
	}
	count, err := CountTaxonomyUnitArticleUsage(db, unit)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该分类单元仍被 %d 篇文章使用，请先调整对应文章的年级/册次/单元后再删除", count)
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

// SettingKeyChannelsFinderUserName 全站视频号 ID（finder-user-name），只开这一键。
const SettingKeyChannelsFinderUserName = "channels_finder_user_name"

func (v Video) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"title":      v.Title,
		"feed_id":    v.FeedID,
		"grade":      v.Grade,
		"semester":   v.Semester,
		"article_id": v.ArticleID,
		"essay_id":   v.EssayID,
		"sort_order": v.SortOrder,
		"status":     v.Status,
	}
}

// FindVideoByID 按 id 查找讲解视频；不存在时返回 gorm.ErrRecordNotFound。
func FindVideoByID(db *gorm.DB, id uint) (Video, error) {
	var item Video
	err := db.First(&item, id).Error
	return item, err
}

// FeedIDTaken 检查 feed_id 是否已被占用；创建传 excludeID=0，更新传自身 id。
func FeedIDTaken(db *gorm.DB, feedID string, excludeID uint) (bool, error) {
	q := db.Model(&Video{}).Where("feed_id = ?", feedID)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckVideoDeletable 物理删除前校验：须已下架（上架中不可删）。
func CheckVideoDeletable(v Video) error {
	if v.Status != StatusDisabled {
		return fmt.Errorf("仅已下架的讲解视频可删除，请先将视频下架")
	}
	return nil
}

// CheckVideoArticleLink 关联专题须存在且已发布；未关联（nil/0）时通过。
func CheckVideoArticleLink(db *gorm.DB, articleID *uint) error {
	if articleID == nil || *articleID == 0 {
		return nil
	}
	var article Article
	if err := db.Select("id", "status").First(&article, *articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("关联专题不存在")
		}
		return err
	}
	if article.Status != ArticleStatusPublished {
		return fmt.Errorf("关联专题未发布")
	}
	return nil
}

// CheckVideoEssayLink 关联优秀作文须存在且已启用；未关联（nil/0）时通过。
func CheckVideoEssayLink(db *gorm.DB, essayID *uint) error {
	if essayID == nil || *essayID == 0 {
		return nil
	}
	var essay EssayExample
	if err := db.Select("id", "status").First(&essay, *essayID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("关联优秀作文不存在")
		}
		return err
	}
	if essay.Status != StatusEnabled {
		return fmt.Errorf("关联优秀作文未启用")
	}
	return nil
}

// GetSystemSetting 按 key 读取配置值；无记录返回空串，不返回 ErrRecordNotFound。
func GetSystemSetting(db *gorm.DB, key string) (string, error) {
	var item SystemSetting
	err := db.Where("setting_key = ?", key).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return item.SettingValue, nil
}

// UpsertSystemSetting 按 setting_key 唯一索引写入；已存在则更新 value / description。
func UpsertSystemSetting(db *gorm.DB, key, value, desc string) error {
	var item SystemSetting
	err := db.Where("setting_key = ?", key).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&SystemSetting{
			SettingKey:   key,
			SettingValue: value,
			Description:  desc,
		}).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&item).Updates(map[string]interface{}{
		"setting_value": value,
		"description":   desc,
	}).Error
}

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

// CheckHomeBlockDeletable 物理删除前校验：须已禁用（启用中不可删，须先禁用）。
func CheckHomeBlockDeletable(block HomeBlock) error {
	if block.Status != StatusDisabled {
		return fmt.Errorf("仅已禁用的首页入口可删除，请先将入口禁用")
	}
	return nil
}

func (k HotSearchKeyword) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"keyword":    k.Keyword,
		"sort_order": k.SortOrder,
		"status":     k.Status,
	}
}

// FindHotSearchKeywordByID 按 id 查找搜索热门词；不存在时返回 gorm.ErrRecordNotFound。
func FindHotSearchKeywordByID(db *gorm.DB, id uint) (HotSearchKeyword, error) {
	var item HotSearchKeyword
	err := db.First(&item, id).Error
	return item, err
}

// CheckHotSearchKeywordDuplicate 检查 keyword 是否与其他记录冲突。
func CheckHotSearchKeywordDuplicate(db *gorm.DB, keyword string, excludeID uint) error {
	q := db.Model(&HotSearchKeyword{}).Where("keyword = ?", keyword)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("hot search keyword already exists")
	}
	return nil
}

// CheckHotSearchKeywordDeletable 物理删除前校验：须已禁用（启用中不可删，须先禁用）。
func CheckHotSearchKeywordDeletable(item HotSearchKeyword) error {
	if item.Status != StatusDisabled {
		return fmt.Errorf("仅已禁用的搜索热门词可删除，请先将热门词禁用")
	}
	return nil
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

// FindAdSlotByID 按 id 查找广告位；不存在时返回 gorm.ErrRecordNotFound。
func FindAdSlotByID(db *gorm.DB, id uint) (AdSlot, error) {
	var slot AdSlot
	err := db.First(&slot, id).Error
	return slot, err
}

// CheckAdSlotDeletable 物理删除前校验：须已禁用（启用中不可删，须先禁用）。
func CheckAdSlotDeletable(slot AdSlot) error {
	if slot.Status != StatusDisabled {
		return fmt.Errorf("仅已禁用的广告位可删除，请先将广告位禁用")
	}
	return nil
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
		"username":    c.Username,
		"status":      c.Status,
	}
}

func FindOfficialAccountConfigByScene(db *gorm.DB, scene string) (OfficialAccountConfig, error) {
	var cfg OfficialAccountConfig
	err := db.Where("scene = ?", scene).First(&cfg).Error
	return cfg, err
}

// FindOfficialAccountConfigByID 按 id 查找公众号导流配置；不存在时返回 gorm.ErrRecordNotFound。
func FindOfficialAccountConfigByID(db *gorm.DB, id uint) (OfficialAccountConfig, error) {
	var cfg OfficialAccountConfig
	err := db.First(&cfg, id).Error
	return cfg, err
}

// CheckOfficialAccountConfigDeletable 物理删除前校验：须已禁用（启用中不可删，须先禁用）。
func CheckOfficialAccountConfigDeletable(cfg OfficialAccountConfig) error {
	if cfg.Status != StatusDisabled {
		return fmt.Errorf("仅已禁用的公众号导流配置可删除，请先将配置禁用")
	}
	return nil
}
