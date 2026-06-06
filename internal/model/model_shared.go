// 【共用】后台维护、小程序读取的内容与运营配置表（一期）。
package model

import (
	"fmt"
	"strings"

	"course_server/internal/pkg/jsontime"
)

// --- 分类与标签（Day4）---

// TaxonomyUnit 教材分类单元。
type TaxonomyUnit struct {
	ID              uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	TextbookVersion string    `gorm:"size:40;index;not null;default:统编版;comment:教材版本" json:"textbookVersion"`
	Grade           int       `gorm:"index;not null;comment:年级(数字)" json:"grade"`
	Semester        string    `gorm:"size:20;index;not null;comment:学期" json:"semester"`
	Unit            int       `gorm:"index;not null;comment:单元序号" json:"unit"`
	UnitTitle       string    `gorm:"size:120;comment:单元标题" json:"unitTitle"`
	SortOrder       int       `gorm:"index;not null;default:0;comment:排序权重" json:"sortOrder"`
	Status          int       `gorm:"index;not null;default:1;comment:状态 0禁用 1启用" json:"status"`
	CreatedAt       jsontime.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt       jsontime.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// ValidTagType 校验标签类型是否为 PRD 定义的四种之一。
func ValidTagType(typ string) bool {
	switch typ {
	case TagTypeArticle, TagTypeEssay, TagTypeDownload, TagTypeTopic:
		return true
	default:
		return false
	}
}

// Tag 标签字典。
type Tag struct {
	ID        uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	Name      string    `gorm:"size:80;not null;index:idx_tag_name_type,unique;comment:标签名" json:"name"`
	Type      string    `gorm:"size:40;not null;index:idx_tag_name_type,unique;comment:标签类型" json:"type"`
	SortOrder int       `gorm:"index;not null;default:0;comment:排序权重" json:"sortOrder"`
	Status    int       `gorm:"index;not null;default:1;comment:状态 0禁用 1启用" json:"status"`
	CreatedAt jsontime.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt jsontime.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// --- 作文专题（Day5）---

// Article 作文专题文章。
type Article struct {
	ID              uint           `gorm:"primaryKey;comment:主键ID" json:"id"`
	Title           string         `gorm:"size:255;index;not null;comment:标题" json:"title"`
	Slug            string         `gorm:"size:255;uniqueIndex;comment:URL别名(唯一)" json:"slug"`
	TextbookVersion string         `gorm:"size:40;index;comment:教材版本" json:"textbookVersion"`
	Grade           int            `gorm:"index;comment:年级" json:"grade"`
	Semester        string         `gorm:"size:20;index;comment:学期" json:"semester"`
	Unit            int            `gorm:"index;comment:单元" json:"unit"`
	EssayTopic      string         `gorm:"size:255;index;comment:作文题目" json:"essayTopic"`
	ContentType     string         `gorm:"size:40;index;comment:内容类型" json:"contentType"`
	Audience        string         `gorm:"size:120;comment:适用读者" json:"audience"`
	Summary         string         `gorm:"size:500;comment:摘要" json:"summary"`
	CoverImg        string         `gorm:"size:500;comment:封面图URL" json:"coverImg"`
	BodyImages      []string       `gorm:"serializer:json;type:json;comment:正文图片URL列表(JSON)" json:"bodyImages,omitempty"`
	ViewAccessType  string         `gorm:"size:40;index;not null;default:free;comment:查看权限 free(一期固定)/member_only/purchase_required(二期预留)" json:"viewAccessType"`
	Status          int            `gorm:"index;not null;default:0;comment:状态 0草稿 1已发布" json:"status"`
	IsRecommended   bool           `gorm:"index;not null;default:false;comment:是否推荐" json:"isRecommended"`
	IsHot           bool           `gorm:"index;not null;default:false;comment:是否热门" json:"isHot"`
	SortOrder       int            `gorm:"index;not null;default:0;comment:排序权重" json:"sortOrder"`
	ViewCount       int64          `gorm:"not null;default:0;comment:浏览次数(冗余计数)" json:"viewCount"`
	ShareCount      int64          `gorm:"not null;default:0;comment:分享次数(冗余计数)" json:"shareCount"`
	FavoriteCount   int64          `gorm:"not null;default:0;comment:收藏次数(冗余计数)" json:"favoriteCount"`
	DownloadCount   int64          `gorm:"not null;default:0;comment:下载次数(冗余计数)" json:"downloadCount"`
	PublishedAt     *jsontime.Time `gorm:"index;comment:发布时间" json:"publishedAt"`
	CreatedAt       jsontime.Time      `gorm:"index;comment:创建时间" json:"createdAt"`
	UpdatedAt       jsontime.Time      `gorm:"comment:更新时间" json:"updatedAt"`
	Essays          []EssayExample `gorm:"foreignKey:ArticleID" json:"essays,omitempty"`
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

// ArticleTag 文章-标签关联；关联 CRUD API 待实现。
type ArticleTag struct {
	ID        uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	ArticleID uint      `gorm:"uniqueIndex:idx_article_tag;not null;comment:文章ID" json:"articleId"`
	TagID     uint      `gorm:"uniqueIndex:idx_article_tag;not null;comment:标签ID" json:"tagId"`
	CreatedAt jsontime.Time `gorm:"comment:创建时间" json:"createdAt"`
}

// --- 优秀作文（Day6）---

// EssayExample 优秀范文。
type EssayExample struct {
	ID            uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	ArticleID     uint      `gorm:"index;comment:所属专题文章ID" json:"articleId"`
	Title         string    `gorm:"size:255;index;not null;comment:范文标题" json:"title"`
	Grade         int       `gorm:"index;comment:适用年级" json:"grade"`
	WordCount     int       `gorm:"comment:字数" json:"wordCount"`
	Body          string    `gorm:"type:longtext;comment:正文" json:"body"`
	Comment       string    `gorm:"type:text;comment:点评/批注" json:"comment"`
	Status        int       `gorm:"index;not null;default:1;comment:状态 0禁用 1启用" json:"status"`
	SortOrder     int       `gorm:"index;not null;default:0;comment:排序权重" json:"sortOrder"`
	ViewCount     int64     `gorm:"not null;default:0;comment:浏览次数(冗余计数)" json:"viewCount"`
	FavoriteCount int64     `gorm:"not null;default:0;comment:收藏次数(冗余计数)" json:"favoriteCount"`
	CreatedAt     jsontime.Time `gorm:"index;comment:创建时间" json:"createdAt"`
	UpdatedAt     jsontime.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// EssayExampleTag 范文-标签关联；关联 CRUD API 待实现。
type EssayExampleTag struct {
	ID        uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	EssayID   uint      `gorm:"uniqueIndex:idx_essay_tag;not null;comment:范文ID" json:"essayId"`
	TagID     uint      `gorm:"uniqueIndex:idx_essay_tag;not null;comment:标签ID" json:"tagId"`
	CreatedAt jsontime.Time `gorm:"comment:创建时间" json:"createdAt"`
}

// --- 下载资料（Day7）---

// DownloadFile 下载资料；access_type/owner 字段二期接订单权限。
type DownloadFile struct {
	ID            uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	OwnerType     string    `gorm:"size:40;index;not null;comment:归属类型 article等" json:"ownerType"`
	OwnerID       uint      `gorm:"index;not null;comment:归属资源ID" json:"ownerId"`
	FileName      string    `gorm:"size:255;not null;comment:展示文件名" json:"fileName"`
	Category      string    `gorm:"size:40;index;not null;comment:资料分类 free/courseware等" json:"category"`
	Tags          string    `gorm:"size:500;comment:标签串(含免费等)" json:"tags"`
	AccessType    string    `gorm:"size:40;index;not null;comment:访问权限 free/purchase_required等" json:"accessType"`
	FileType      string    `gorm:"size:40;comment:文件类型 pdf等" json:"fileType"`
	FileURL       string    `gorm:"size:1000;not null;comment:文件存储URL" json:"fileUrl"`
	FileSize      int64     `gorm:"comment:文件大小(字节)" json:"fileSize"`
	NeedLogin     bool      `gorm:"not null;default:true;comment:是否需登录下载" json:"needLogin"`
	DownloadCount int64     `gorm:"not null;default:0;comment:下载次数(冗余计数)" json:"downloadCount"`
	SortOrder     int       `gorm:"index;not null;default:0;comment:排序权重" json:"sortOrder"`
	Status        int       `gorm:"index;not null;default:1;comment:状态 0禁用 1启用" json:"status"`
	CreatedAt     jsontime.Time `gorm:"index;comment:创建时间" json:"createdAt"`
	UpdatedAt     jsontime.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// --- 运营配置（后台可配，小程序读取）---

// AdSlot 广告位配置。
type AdSlot struct {
	ID        uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	SlotKey   string    `gorm:"size:80;uniqueIndex;not null;comment:广告位唯一键" json:"slotKey"`
	SlotName  string    `gorm:"size:120;not null;comment:广告位名称" json:"slotName"`
	Page      string    `gorm:"size:80;index;not null;comment:所属页面" json:"page"`
	Position  string    `gorm:"size:80;index;comment:页面内位置" json:"position"`
	AdUnitID  string    `gorm:"size:120;comment:微信广告单元ID" json:"adUnitId"`
	Status    int       `gorm:"index;not null;default:1;comment:状态 0禁用 1启用" json:"status"`
	SortOrder int       `gorm:"index;not null;default:0;comment:排序权重" json:"sortOrder"`
	CreatedAt jsontime.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt jsontime.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// HomeBlock 首页区块配置。
type HomeBlock struct {
	ID          uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	BlockKey    string    `gorm:"size:80;uniqueIndex;not null;comment:区块唯一键" json:"blockKey"`
	Title       string    `gorm:"size:120;not null;comment:区块标题" json:"title"`
	ContentType string    `gorm:"size:40;index;not null;comment:内容类型" json:"contentType"`
	ConfigJSON  string    `gorm:"type:text;comment:区块配置JSON" json:"configJson"`
	SortOrder   int       `gorm:"index;not null;default:0;comment:排序权重" json:"sortOrder"`
	Status      int       `gorm:"index;not null;default:1;comment:状态 0禁用 1启用" json:"status"`
	CreatedAt   jsontime.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt   jsontime.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// OfficialAccountConfig 公众号导流配置。
type OfficialAccountConfig struct {
	ID          uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	Scene       string    `gorm:"size:80;uniqueIndex;not null;comment:场景标识" json:"scene"`
	Title       string    `gorm:"size:120;comment:展示标题" json:"title"`
	Description string    `gorm:"size:500;comment:描述文案" json:"description"`
	QrcodeURL   string    `gorm:"size:500;comment:二维码图片URL" json:"qrcodeUrl"`
	LinkText    string    `gorm:"size:120;comment:引导链接文案" json:"linkText"`
	Status      int       `gorm:"index;not null;default:1;comment:状态 0禁用 1启用" json:"status"`
	CreatedAt   jsontime.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt   jsontime.Time `gorm:"comment:更新时间" json:"updatedAt"`
}
