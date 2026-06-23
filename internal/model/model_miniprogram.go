// 【小程序】微信用户与端上行为日志表（一期）。
package model

import "course_server/internal/pkg/jsontime"

// User 微信用户；OpenID 唯一，UnionID 预留多端打通。
type User struct {
	ID          uint          `gorm:"primaryKey;comment:主键ID" json:"id"`
	OpenID      string        `gorm:"size:128;uniqueIndex;not null;comment:微信OpenID" json:"openid"`
	UnionID     string        `gorm:"size:128;index;comment:微信UnionID(预留)" json:"unionid"`
	Nickname    string        `gorm:"size:80;comment:昵称" json:"nickname"`
	Avatar      string        `gorm:"size:500;comment:头像URL" json:"avatar"`
	SourceScene string        `gorm:"size:80;index;comment:注册/访问场景" json:"sourceScene"`
	Status      int           `gorm:"not null;default:1;comment:状态 0禁用 1正常" json:"status"`
	BanReason   string        `gorm:"size:255;comment:封禁原因" json:"banReason"`
	CreatedAt   jsontime.Time `gorm:"index;comment:注册时间" json:"createdAt"`
	UpdatedAt   jsontime.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// DownloadLog 用户下载行为日志。
type DownloadLog struct {
	ID          uint          `gorm:"primaryKey;comment:主键ID" json:"id"`
	UserID      uint          `gorm:"index;comment:小程序用户ID" json:"userId"`
	FileID      uint          `gorm:"index;not null;comment:下载资料ID" json:"fileId"`
	OwnerType   string        `gorm:"size:40;index;comment:归属类型快照" json:"ownerType"`
	OwnerID     uint          `gorm:"index;comment:归属ID快照" json:"ownerId"`
	AccessType  string        `gorm:"size:40;index;comment:权限类型快照" json:"accessType"`
	SourceScene string        `gorm:"size:80;index;comment:来源场景" json:"sourceScene"`
	CreatedAt   jsontime.Time `gorm:"index;comment:下载时间" json:"createdAt"`
}

// Favorite 用户收藏。
type Favorite struct {
	ID         uint          `gorm:"primaryKey;comment:主键ID" json:"id"`
	UserID     uint          `gorm:"uniqueIndex:idx_user_fav;not null;comment:用户ID" json:"userId"`
	TargetType string        `gorm:"size:40;uniqueIndex:idx_user_fav;not null;comment:收藏目标类型 article/essay" json:"targetType"`
	TargetID   uint          `gorm:"uniqueIndex:idx_user_fav;not null;comment:收藏目标ID" json:"targetId"`
	CreatedAt  jsontime.Time `gorm:"index;comment:收藏时间" json:"createdAt"`
}

// ViewLog 浏览日志。
type ViewLog struct {
	ID          uint          `gorm:"primaryKey;comment:主键ID" json:"id"`
	UserID      uint          `gorm:"index;comment:用户ID(可匿名为0)" json:"userId"`
	TargetType  string        `gorm:"size:40;index;comment:内容类型" json:"targetType"`
	TargetID    uint          `gorm:"index;comment:内容ID" json:"targetId"`
	PagePath    string        `gorm:"size:255;index;comment:小程序页面路径" json:"pagePath"`
	SourceScene string        `gorm:"size:80;index;comment:来源场景" json:"sourceScene"`
	CreatedAt   jsontime.Time `gorm:"index;comment:浏览时间" json:"createdAt"`
}

// SearchLog 搜索日志。
type SearchLog struct {
	ID          uint          `gorm:"primaryKey;comment:主键ID" json:"id"`
	UserID      uint          `gorm:"index;comment:用户ID" json:"userId"`
	Keyword     string        `gorm:"size:120;index;not null;comment:搜索关键词" json:"keyword"`
	ResultCount int           `gorm:"comment:命中结果数" json:"resultCount"`
	SourceScene string        `gorm:"size:80;index;comment:来源场景" json:"sourceScene"`
	CreatedAt   jsontime.Time `gorm:"index;comment:搜索时间" json:"createdAt"`
}

// ShareLog 分享日志。
type ShareLog struct {
	ID           uint          `gorm:"primaryKey;comment:主键ID" json:"id"`
	UserID       uint          `gorm:"index;comment:用户ID" json:"userId"`
	TargetType   string        `gorm:"size:40;index;comment:分享内容类型" json:"targetType"`
	TargetID     uint          `gorm:"index;comment:分享内容ID" json:"targetId"`
	ShareChannel string        `gorm:"size:40;index;comment:分享渠道" json:"shareChannel"`
	SourceScene  string        `gorm:"size:80;index;comment:来源场景" json:"sourceScene"`
	CreatedAt    jsontime.Time `gorm:"index;comment:分享时间" json:"createdAt"`
}

// AdEvent 广告行为日志。
type AdEvent struct {
	ID         uint          `gorm:"primaryKey;comment:主键ID" json:"id"`
	UserID     uint          `gorm:"index;comment:用户ID" json:"userId"`
	SlotKey    string        `gorm:"size:80;index;not null;comment:广告位键" json:"slotKey"`
	EventType  string        `gorm:"size:40;index;not null;comment:事件类型 expose/click等" json:"eventType"`
	PagePath   string        `gorm:"size:255;index;comment:页面路径" json:"pagePath"`
	TargetType string        `gorm:"size:40;index;comment:关联内容类型" json:"targetType"`
	TargetID   uint          `gorm:"index;comment:关联内容ID" json:"targetId"`
	CreatedAt  jsontime.Time `gorm:"index;comment:事件时间" json:"createdAt"`
}

// Feedback 用户意见反馈。
type Feedback struct {
	ID        uint           `gorm:"primaryKey;comment:主键ID" json:"id"`
	UserID    uint           `gorm:"index;not null;comment:提交用户ID" json:"userId"`
	Content   string         `gorm:"type:text;not null;comment:反馈正文" json:"content"`
	Status    string         `gorm:"size:20;index;not null;default:pending;comment:状态 pending/replied" json:"status"`
	Reply     string         `gorm:"type:text;comment:管理员回复" json:"reply"`
	RepliedAt *jsontime.Time `gorm:"comment:回复时间" json:"repliedAt"`
	RepliedBy uint           `gorm:"index;comment:回复管理员ID" json:"repliedBy"`
	CreatedAt jsontime.Time  `gorm:"index;comment:提交时间" json:"createdAt"`
	UpdatedAt jsontime.Time  `gorm:"comment:更新时间" json:"updatedAt"`
}
