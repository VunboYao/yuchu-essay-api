// Package adunlock 签发与校验下载激励视频解锁凭证（短时 JWT，不依赖 Redis）。
package adunlock

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultExpiresIn = 300 // 秒，与 sign-url 伪签名有效期一致

// Claims 下载解锁 token 载荷；type 固定 ad_unlock 与登录 JWT 区分。
type Claims struct {
	FileID  uint   `json:"fileId"`
	UserID  uint   `json:"userId"`
	SlotKey string `json:"slotKey"`
	Type    string `json:"type"`
	jwt.RegisteredClaims
}

// Sign 签发下载解锁 token，返回 token 与 expiresIn（秒）。
func Sign(secret, issuer string, fileID, userID uint, slotKey string) (string, int, error) {
	if secret == "" {
		return "", 0, errors.New("secret is empty")
	}
	if fileID == 0 {
		return "", 0, errors.New("fileId is required")
	}
	if slotKey == "" {
		return "", 0, errors.New("slotKey is required")
	}
	now := time.Now()
	claims := Claims{
		FileID:  fileID,
		UserID:  userID,
		SlotKey: slotKey,
		Type:    "ad_unlock",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultExpiresIn * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return token, defaultExpiresIn, nil
}

// Verify 校验 unlockToken 是否匹配指定 fileID 且未过期。
func Verify(secret, tokenString string, fileID uint) (*Claims, error) {
	if secret == "" {
		return nil, errors.New("secret is empty")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid unlock token")
	}
	if claims.Type != "ad_unlock" {
		return nil, errors.New("invalid unlock token type")
	}
	if claims.FileID != fileID {
		return nil, errors.New("unlock token file mismatch")
	}
	return claims, nil
}
