// Package config 从 config.yaml 加载运行时配置（Day1）。
// OSS、Wechat 字段供 Day7/Day8 接入上传与微信登录时使用。
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`   // HTTP 服务监听地址配置
	Database DatabaseConfig `yaml:"database"` // MySQL 数据库连接配置
	Redis    RedisConfig    `yaml:"redis"`    // Redis 缓存连接配置
	JWT      JWTConfig      `yaml:"jwt"`      // JWT 签名与签发者配置
	OSS      OSSConfig      `yaml:"oss"`      // 对象存储（OSS）配置，Day7 接入上传时使用
	Wechat   WechatConfig   `yaml:"wechat"`   // 微信小程序配置，Day8 接入登录时使用
}

// ServerConfig HTTP 服务监听配置
type ServerConfig struct {
	Addr string `yaml:"addr"` // 监听地址，如 ":8080"
}

// DatabaseConfig MySQL 数据库连接配置
type DatabaseConfig struct {
	DSN string `yaml:"dsn"` // 数据库连接字符串（user:pass@tcp(host:port)/dbname）
}

// RedisConfig Redis 缓存连接配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`     // Redis 服务地址，如 "127.0.0.1:6379"
	Password string `yaml:"password"` // Redis 认证密码，无密码时留空
	DB       int    `yaml:"db"`       // 使用的 Redis 数据库编号（默认 0）
}

// JWTConfig JWT 令牌签发与验证配置
type JWTConfig struct {
	Secret string `yaml:"secret"` // JWT 签名密钥，生产环境必须替换
	Issuer string `yaml:"issuer"` // JWT 签发者标识
}

// OSSConfig 阿里云对象存储配置，Day7 接入文件上传时使用
type OSSConfig struct {
	Endpoint        string `yaml:"endpoint"`          // OSS 服务端点，如 "oss-cn-hangzhou.aliyuncs.com"
	Bucket          string `yaml:"bucket"`            // 存储桶名称
	AccessKeyID     string `yaml:"access_key_id"`     // 阿里云 AccessKey ID
	AccessKeySecret string `yaml:"access_key_secret"` // 阿里云 AccessKey Secret
}

// WechatConfig 微信小程序配置，Day8 接入微信登录时使用
type WechatConfig struct {
	AppID     string `yaml:"app_id"`     // 微信小程序 AppID
	AppSecret string `yaml:"app_secret"` // 微信小程序 AppSecret
}

// Load 读取 YAML 配置；文件不存在时返回 defaultConfig，便于本地快速启动。
func Load(path string) (*Config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// defaultConfig 返回默认配置，便于本地快速启动。
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{Addr: ":8080"},
		JWT: JWTConfig{
			Secret: "change-me-in-production",
			Issuer: "essay-mvp",
		},
	}
}
