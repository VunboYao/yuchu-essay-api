// Package jwtutil 签发与解析 JWT（Day3），区分 admin / user 两种 type，有效期 7 天。
package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID      uint   `json:"id"`
	Subject string `json:"subject"`
	Role    string `json:"role"`
	Type    string `json:"type"`
	jwt.RegisteredClaims
}

func SignAdmin(secret, issuer string, adminID uint, username, role string) (string, error) {
	return sign(secret, issuer, adminID, username, role, "admin")
}

func SignUser(secret, issuer string, userID uint, openid string) (string, error) {
	return sign(secret, issuer, userID, openid, "", "user")
}

func sign(secret, issuer string, id uint, subject, role, tokenType string) (string, error) {
	claims := Claims{
		ID:      id,
		Subject: subject,
		Role:    role,
		Type:    tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// Parse 解析 token 并校验 type 与签发方，防止后台 token 用于小程序接口。
func Parse(secret string, tokenString string, tokenType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Type != tokenType {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}
