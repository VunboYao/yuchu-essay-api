// 【后台】管理员、审计与系统配置表（一期）。
package model

import (
	"strings"

	"course_server/internal/pkg/jsontime"
)

// 后台操作审计目标类型（与 Admin 前端筛选项一致；article/essay/download_file 复用 models.go 常量）。
const (
	AdminTargetTypeTag             = "tag"
	AdminTargetTypeTaxonomyUnit    = "taxonomy_unit"
	AdminTargetTypeHomeBlock       = "home_block"
	AdminTargetTypeVideo           = "video"
	AdminTargetTypeVideoChannel    = "video_channel"
	AdminTargetTypeAdSlot          = "ad_slot"
	AdminTargetTypeOfficialAccount = "official_account"
	AdminTargetTypeUpload          = "upload"
)

// adminPathToTargetType 将 /admin/api/v1/<resource> 首段映射为标准 targetType。
var adminPathToTargetType = map[string]string{
	"articles":                 TargetTypeArticle,
	"essay-examples":           TargetTypeEssay,
	"download-files":           TargetTypeDownloadFile,
	"tags":                     AdminTargetTypeTag,
	"taxonomy-units":           AdminTargetTypeTaxonomyUnit,
	"home-blocks":              AdminTargetTypeHomeBlock,
	"videos":                   AdminTargetTypeVideo,
	"video-channel":            AdminTargetTypeVideoChannel,
	"ad-slots":                 AdminTargetTypeAdSlot,
	"official-account-configs": AdminTargetTypeOfficialAccount,
	"upload":                   AdminTargetTypeUpload,
}

// AdminOpLogTargetTypeFromPath 从 Gin FullPath 解析并映射为标准审计目标类型。
func AdminOpLogTargetTypeFromPath(fullPath string) string {
	const prefix = "/admin/api/v1/"
	if !strings.HasPrefix(fullPath, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(fullPath, prefix)
	resource := rest
	if idx := strings.IndexByte(rest, '/'); idx >= 0 {
		resource = rest[:idx]
	}
	if t, ok := adminPathToTargetType[resource]; ok {
		return t
	}
	return resource
}

// Admin 管理员账号。
type Admin struct {
	ID           uint       `gorm:"primaryKey;comment:主键ID" json:"id"`
	Username     string     `gorm:"size:64;uniqueIndex;not null;comment:登录用户名" json:"username"`
	PasswordHash string     `gorm:"size:255;not null;comment:密码哈希(bcrypt)" json:"-"`
	Role         string     `gorm:"size:32;not null;default:admin;comment:角色标识" json:"role"`
	Status       int        `gorm:"not null;default:1;comment:状态 0禁用 1启用" json:"status"`
	LastLoginAt  *jsontime.Time `gorm:"comment:最近登录时间" json:"lastLoginAt"`
	CreatedAt    jsontime.Time  `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt    jsontime.Time  `gorm:"comment:更新时间" json:"updatedAt"`
}

// AdminOperationLog 操作审计；由 middleware.AdminOperationLog 写入。
type AdminOperationLog struct {
	ID         uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	AdminID    uint      `gorm:"index;not null;comment:操作管理员ID" json:"adminId"`
	Action     string    `gorm:"size:80;index;not null;comment:操作动作编码" json:"action"`
	TargetType string    `gorm:"size:40;index;comment:目标资源类型" json:"targetType"`
	TargetID   uint      `gorm:"index;comment:目标资源ID" json:"targetId"`
	Remark     string    `gorm:"type:text;comment:备注说明" json:"remark"`
	CreatedAt  jsontime.Time `gorm:"index;comment:操作时间" json:"createdAt"`
}

// SystemSetting 系统 KV 配置；读写 API 待实现。
type SystemSetting struct {
	ID           uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	SettingKey   string    `gorm:"size:120;uniqueIndex;not null;comment:配置键" json:"settingKey"`
	SettingValue string    `gorm:"type:text;comment:配置值" json:"settingValue"`
	Description  string    `gorm:"size:255;comment:配置说明" json:"description"`
	CreatedAt    jsontime.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt    jsontime.Time `gorm:"comment:更新时间" json:"updatedAt"`
}
