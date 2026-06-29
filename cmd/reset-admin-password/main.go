package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	"course_server/internal/config"
	"course_server/internal/database"
	"course_server/internal/model"

	"golang.org/x/crypto/bcrypt"
)

const adminPasswordEnv = "ADMIN_PASSWORD"

func main() {
	configPath := flag.String("config", "config.yaml", "生产 config.yaml 路径")
	username := flag.String("username", "admin", "要重置密码的管理员用户名")
	flag.Parse()

	password := strings.TrimSpace(os.Getenv(adminPasswordEnv))
	if err := validatePassword(password); err != nil {
		log.Fatalf("invalid %s: %v", adminPasswordEnv, err)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := database.InitMySQL(cfg.Database.DSN); err != nil {
		log.Fatalf("init mysql: %v", err)
	}
	defer func() {
		if err := database.CloseMySQL(); err != nil {
			log.Printf("close mysql: %v", err)
		}
	}()

	// 只更新指定管理员的 bcrypt 哈希，不把明文密码写入日志或配置文件。
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	result := database.DB.Model(&model.Admin{}).
		Where("username = ?", strings.TrimSpace(*username)).
		Update("password_hash", string(hash))
	if result.Error != nil {
		log.Fatalf("update admin password: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		log.Fatalf("admin %q not found", strings.TrimSpace(*username))
	}

	fmt.Printf("管理员 %s 密码已重置，请使用新密码登录。\n", strings.TrimSpace(*username))
}

func validatePassword(password string) error {
	if len(password) < 16 {
		return fmt.Errorf("长度至少 16 个字符")
	}

	var hasLower, hasUpper, hasDigit, hasSymbol bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	if !hasLower || !hasUpper || !hasDigit || !hasSymbol {
		return fmt.Errorf("需同时包含大小写字母、数字和符号")
	}
	return nil
}
