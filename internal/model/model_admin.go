// 【后台】管理员、审计与系统配置表（一期）。
package model

import "time"

// Admin 管理员账号。
type Admin struct {
	ID           uint       `gorm:"primaryKey;comment:主键ID" json:"id"`
	Username     string     `gorm:"size:64;uniqueIndex;not null;comment:登录用户名" json:"username"`
	PasswordHash string     `gorm:"size:255;not null;comment:密码哈希(bcrypt)" json:"-"`
	Role         string     `gorm:"size:32;not null;default:admin;comment:角色标识" json:"role"`
	Status       int        `gorm:"not null;default:1;comment:状态 0禁用 1启用" json:"status"`
	LastLoginAt  *time.Time `gorm:"comment:最近登录时间" json:"lastLoginAt"`
	CreatedAt    time.Time  `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"comment:更新时间" json:"updatedAt"`
}

// AdminOperationLog 操作审计；由 middleware.AdminOperationLog 写入。
type AdminOperationLog struct {
	ID         uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	AdminID    uint      `gorm:"index;not null;comment:操作管理员ID" json:"adminId"`
	Action     string    `gorm:"size:80;index;not null;comment:操作动作编码" json:"action"`
	TargetType string    `gorm:"size:40;index;comment:目标资源类型" json:"targetType"`
	TargetID   uint      `gorm:"index;comment:目标资源ID" json:"targetId"`
	Remark     string    `gorm:"type:text;comment:备注说明" json:"remark"`
	CreatedAt  time.Time `gorm:"index;comment:操作时间" json:"createdAt"`
}

// SystemSetting 系统 KV 配置；读写 API 待实现。
type SystemSetting struct {
	ID           uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	SettingKey   string    `gorm:"size:120;uniqueIndex;not null;comment:配置键" json:"settingKey"`
	SettingValue string    `gorm:"type:text;comment:配置值" json:"settingValue"`
	Description  string    `gorm:"size:255;comment:配置说明" json:"description"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"comment:更新时间" json:"updatedAt"`
}
