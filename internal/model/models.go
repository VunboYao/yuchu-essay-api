// Package model 定义一期全部 GORM 模型与业务常量（Day2）。
// 查询字段上的 index 标签对应 PRD 要求：openid、slug、分类筛选、target_type/id、created_at 等。
package model

import (
	"fmt"
	"strings"
	"time"
)

// 通用与内容状态、下载权限常量（禁止在 handler 中散落魔法值）。
const (
	StatusDisabled = 0
	StatusEnabled  = 1

	ArticleStatusDraft     = 0
	ArticleStatusPublished = 1

	DownloadCategoryFree       = "free"
	DownloadCategoryCourseware = "courseware"
	DownloadCategoryMember     = "member"
	DownloadCategoryInternal   = "internal"

	AccessTypeFree             = "free"
	AccessTypePurchaseRequired = "purchase_required"
	AccessTypeMemberOnly       = "member_only"
	AccessTypeInternal         = "internal" // 一期一律拒绝公开下载签名

	// 收藏/分享/浏览等行为的目标类型，用于计数回写与日志归类。
	TargetTypeArticle = "article"
	TargetTypeEssay   = "essay"

	MaxBodyImages = 20
)

// --- 鉴权与用户（Day3 / Day8）---

type Admin struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Role         string     `gorm:"size:32;not null;default:admin" json:"role"`
	Status       int        `gorm:"not null;default:1" json:"status"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// AdminOperationLog 后台操作审计表（Day3）；由 middleware.AdminOperationLog 写入。
type AdminOperationLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	AdminID    uint      `gorm:"index;not null" json:"adminId"`
	Action     string    `gorm:"size:80;index;not null" json:"action"`
	TargetType string    `gorm:"size:40;index" json:"targetType"`
	TargetID   uint      `gorm:"index" json:"targetId"`
	Remark     string    `gorm:"type:text" json:"remark"`
	CreatedAt  time.Time `gorm:"index" json:"createdAt"`
}

// User 小程序用户；OpenID 唯一，UnionID 预留供多端打通。
type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OpenID      string    `gorm:"size:128;uniqueIndex;not null" json:"openid"`
	UnionID     string    `gorm:"size:128;index" json:"unionid"`
	Nickname    string    `gorm:"size:80" json:"nickname"`
	Avatar      string    `gorm:"size:500" json:"avatar"`
	SourceScene string    `gorm:"size:80;index" json:"sourceScene"`
	Status      int       `gorm:"not null;default:1" json:"status"`
	BanReason   string    `gorm:"size:255" json:"banReason"`
	CreatedAt   time.Time `gorm:"index" json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// --- 分类与标签（Day4）---

type TaxonomyUnit struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TextbookVersion string    `gorm:"size:40;index;not null" json:"textbookVersion"`
	Grade           int       `gorm:"index;not null" json:"grade"`
	Semester        string    `gorm:"size:20;index;not null" json:"semester"`
	Unit            int       `gorm:"index;not null" json:"unit"`
	UnitTitle       string    `gorm:"size:120" json:"unitTitle"`
	SortOrder       int       `gorm:"index;not null;default:0" json:"sortOrder"`
	Status          int       `gorm:"index;not null;default:1" json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:80;not null;index:idx_tag_name_type,unique" json:"name"`
	Type      string    `gorm:"size:40;not null;index:idx_tag_name_type,unique" json:"type"`
	SortOrder int       `gorm:"index;not null;default:0" json:"sortOrder"`
	Status    int       `gorm:"index;not null;default:1" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// --- 作文专题（Day5）---

type Article struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	Title           string           `gorm:"size:255;index;not null" json:"title"`
	Slug            string           `gorm:"size:255;uniqueIndex" json:"slug"`
	TextbookVersion string           `gorm:"size:40;index" json:"textbookVersion"`
	Grade           int              `gorm:"index" json:"grade"`
	Semester        string           `gorm:"size:20;index" json:"semester"`
	Unit            int              `gorm:"index" json:"unit"`
	EssayTopic      string           `gorm:"size:255;index" json:"essayTopic"`
	ContentType     string           `gorm:"size:40;index" json:"contentType"`
	Audience        string           `gorm:"size:120" json:"audience"`
	Summary         string           `gorm:"size:500" json:"summary"`
	CoverImg        string           `gorm:"size:500" json:"coverImg"`
	BodyImages      []string         `gorm:"serializer:json;type:json" json:"bodyImages,omitempty"`
	Status          int              `gorm:"index;not null;default:0" json:"status"`
	IsRecommended   bool             `gorm:"index;not null;default:false" json:"isRecommended"`
	IsHot           bool             `gorm:"index;not null;default:false" json:"isHot"`
	SortOrder       int              `gorm:"index;not null;default:0" json:"sortOrder"`
	ViewCount       int64            `gorm:"not null;default:0" json:"viewCount"`
	ShareCount      int64            `gorm:"not null;default:0" json:"shareCount"`
	FavoriteCount   int64            `gorm:"not null;default:0" json:"favoriteCount"`
	DownloadCount   int64            `gorm:"not null;default:0" json:"downloadCount"`
	PublishedAt     *time.Time       `gorm:"index" json:"publishedAt"`
	CreatedAt       time.Time        `gorm:"index" json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
	Essays          []EssayExample   `gorm:"foreignKey:ArticleID" json:"essays,omitempty"`
}

// ValidateBodyImages 校验正文图片列表：最多 MaxBodyImages 张，URL 不可为空。
func ValidateBodyImages(images []string) error {
	if len(images) > MaxBodyImages {
		return fmt.Errorf("bodyImages exceeds max %d", MaxBodyImages)
	}
	for i, url := range images {
		if strings.TrimSpace(url) == "" {
			return fmt.Errorf("bodyImages[%d] is empty", i)
		}
	}
	return nil
}

// ArticleTag 文章-标签多对多（Day5）；表已迁移，关联 CRUD API 待实现。
type ArticleTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ArticleID uint      `gorm:"uniqueIndex:idx_article_tag;not null" json:"articleId"`
	TagID     uint      `gorm:"uniqueIndex:idx_article_tag;not null" json:"tagId"`
	CreatedAt time.Time `json:"createdAt"`
}

// --- 优秀作文（Day6）---

type EssayExample struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ArticleID     uint      `gorm:"index" json:"articleId"`
	Title         string    `gorm:"size:255;index;not null" json:"title"`
	Grade         int       `gorm:"index" json:"grade"`
	WordCount     int       `json:"wordCount"`
	Body          string    `gorm:"type:longtext" json:"body"`
	Comment       string    `gorm:"type:text" json:"comment"`
	Status        int       `gorm:"index;not null;default:1" json:"status"`
	SortOrder     int       `gorm:"index;not null;default:0" json:"sortOrder"`
	ViewCount     int64     `gorm:"not null;default:0" json:"viewCount"`
	FavoriteCount int64     `gorm:"not null;default:0" json:"favoriteCount"`
	CreatedAt     time.Time `gorm:"index" json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// EssayExampleTag 范文-标签多对多（Day6）；表已迁移，关联 CRUD API 待实现。
type EssayExampleTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EssayID   uint      `gorm:"uniqueIndex:idx_essay_tag;not null" json:"essayId"`
	TagID     uint      `gorm:"uniqueIndex:idx_essay_tag;not null" json:"tagId"`
	CreatedAt time.Time `json:"createdAt"`
}

// --- 下载资料（Day7）---

type DownloadFile struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	OwnerType     string    `gorm:"size:40;index;not null" json:"ownerType"`
	OwnerID       uint      `gorm:"index;not null" json:"ownerId"`
	FileName      string    `gorm:"size:255;not null" json:"fileName"`
	Category      string    `gorm:"size:40;index;not null" json:"category"`
	Tags          string    `gorm:"size:500" json:"tags"`
	AccessType    string    `gorm:"size:40;index;not null" json:"accessType"`
	FileType      string    `gorm:"size:40" json:"fileType"`
	FileURL       string    `gorm:"size:1000;not null" json:"fileUrl"`
	FileSize      int64     `json:"fileSize"`
	NeedLogin     bool      `gorm:"not null;default:true" json:"needLogin"`
	DownloadCount int64     `gorm:"not null;default:0" json:"downloadCount"`
	SortOrder     int       `gorm:"index;not null;default:0" json:"sortOrder"`
	Status        int       `gorm:"index;not null;default:1" json:"status"`
	CreatedAt     time.Time `gorm:"index" json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type DownloadLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index" json:"userId"`
	FileID       uint      `gorm:"index;not null" json:"fileId"`
	OwnerType    string    `gorm:"size:40;index" json:"ownerType"`
	OwnerID      uint      `gorm:"index" json:"ownerId"`
	FileCategory string    `gorm:"size:40;index" json:"fileCategory"`
	AccessType   string    `gorm:"size:40;index" json:"accessType"`
	SourceScene  string    `gorm:"size:80;index" json:"sourceScene"`
	CreatedAt    time.Time `gorm:"index" json:"createdAt"`
}

// --- 用户行为与日志（Day10 / Day11）---

type Favorite struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"uniqueIndex:idx_user_fav;not null" json:"userId"`
	TargetType string    `gorm:"size:40;uniqueIndex:idx_user_fav;not null" json:"targetType"`
	TargetID   uint      `gorm:"uniqueIndex:idx_user_fav;not null" json:"targetId"`
	CreatedAt  time.Time `gorm:"index" json:"createdAt"`
}

type ViewLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"userId"`
	TargetType  string    `gorm:"size:40;index" json:"targetType"`
	TargetID    uint      `gorm:"index" json:"targetId"`
	PagePath    string    `gorm:"size:255;index" json:"pagePath"`
	SourceScene string    `gorm:"size:80;index" json:"sourceScene"`
	CreatedAt   time.Time `gorm:"index" json:"createdAt"`
}

type SearchLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"userId"`
	Keyword     string    `gorm:"size:120;index;not null" json:"keyword"`
	ResultCount int       `json:"resultCount"`
	SourceScene string    `gorm:"size:80;index" json:"sourceScene"`
	CreatedAt   time.Time `gorm:"index" json:"createdAt"`
}

type ShareLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index" json:"userId"`
	TargetType   string    `gorm:"size:40;index" json:"targetType"`
	TargetID     uint      `gorm:"index" json:"targetId"`
	ShareChannel string    `gorm:"size:40;index" json:"shareChannel"`
	SourceScene  string    `gorm:"size:80;index" json:"sourceScene"`
	CreatedAt    time.Time `gorm:"index" json:"createdAt"`
}

type AdSlot struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SlotKey   string    `gorm:"size:80;uniqueIndex;not null" json:"slotKey"`
	SlotName  string    `gorm:"size:120;not null" json:"slotName"`
	Page      string    `gorm:"size:80;index;not null" json:"page"`
	Position  string    `gorm:"size:80;index" json:"position"`
	AdUnitID  string    `gorm:"size:120" json:"adUnitId"`
	Status    int       `gorm:"index;not null;default:1" json:"status"`
	SortOrder int       `gorm:"index;not null;default:0" json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AdEvent struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"userId"`
	SlotKey    string    `gorm:"size:80;index;not null" json:"slotKey"`
	EventType  string    `gorm:"size:40;index;not null" json:"eventType"`
	PagePath   string    `gorm:"size:255;index" json:"pagePath"`
	TargetType string    `gorm:"size:40;index" json:"targetType"`
	TargetID   uint      `gorm:"index" json:"targetId"`
	CreatedAt  time.Time `gorm:"index" json:"createdAt"`
}

// --- 运营配置（后台可配，小程序读取）---

type HomeBlock struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	BlockKey    string    `gorm:"size:80;uniqueIndex;not null" json:"blockKey"`
	Title       string    `gorm:"size:120;not null" json:"title"`
	ContentType string    `gorm:"size:40;index;not null" json:"contentType"`
	ConfigJSON  string    `gorm:"type:text" json:"configJson"`
	SortOrder   int       `gorm:"index;not null;default:0" json:"sortOrder"`
	Status      int       `gorm:"index;not null;default:1" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type OfficialAccountConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Scene       string    `gorm:"size:80;uniqueIndex;not null" json:"scene"`
	Title       string    `gorm:"size:120" json:"title"`
	Description string    `gorm:"size:500" json:"description"`
	QrcodeURL   string    `gorm:"size:500" json:"qrcodeUrl"`
	LinkText    string    `gorm:"size:120" json:"linkText"`
	Status      int       `gorm:"index;not null;default:1" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// SystemSetting 系统 KV 配置；表已迁移，读写 API 待实现。
type SystemSetting struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SettingKey   string    `gorm:"size:120;uniqueIndex;not null" json:"settingKey"`
	SettingValue string    `gorm:"type:text" json:"settingValue"`
	Description  string    `gorm:"size:255" json:"description"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
