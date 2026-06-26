// Package wechat 封装微信小程序服务端 API（code2session）。
package wechat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const code2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

// Session 微信 code2session 成功响应。
type Session struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
}

type apiError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Code2Session 用 wx.login 的 code 换取 openid；session_key 仅服务端使用，勿下发前端。
func Code2Session(appID, appSecret, code string) (*Session, error) {
	if appID == "" || appSecret == "" {
		return nil, fmt.Errorf("wechat app_id or app_secret not configured")
	}
	q := url.Values{}
	q.Set("appid", appID)
	q.Set("secret", appSecret)
	q.Set("js_code", code)
	q.Set("grant_type", "authorization_code")
	reqURL := code2SessionURL + "?" + q.Encode()
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("wechat code2session request: %w", err)
	}
	defer resp.Body.Close()
	var out struct {
		Session
		apiError
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("wechat code2session decode: %w", err)
	}
	if out.ErrCode != 0 {
		return nil, fmt.Errorf("wechat code2session errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
	}
	if out.OpenID == "" {
		return nil, fmt.Errorf("wechat code2session empty openid")
	}
	return &out.Session, nil
}

// DevMode 未配置微信凭证时走开发占位登录。
func DevMode(appID, appSecret string) bool {
	return appID == "" || appSecret == ""
}
