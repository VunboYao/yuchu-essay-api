package wechat_test

import (
	"strings"
	"testing"

	"course_server/internal/pkg/wechat"
)

func TestDevMode(t *testing.T) {
	if !wechat.DevMode("", "secret") {
		t.Fatal("expected dev mode when app_id empty")
	}
	if wechat.DevMode("wxid", "secret") {
		t.Fatal("expected prod mode when both set")
	}
}

func TestCode2SessionRequiresConfig(t *testing.T) {
	_, err := wechat.Code2Session("", "", "code")
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("err=%v", err)
	}
}
