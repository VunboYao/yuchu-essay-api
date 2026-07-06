package ossutil_test

import (
	"strings"
	"testing"

	"course_server/internal/config"
	"course_server/internal/pkg/ossutil"
)

func TestObjectKeyFromURL(t *testing.T) {
	cfg := config.OSSConfig{
		Endpoint:  "oss-cn-hangzhou.aliyuncs.com",
		Bucket:    "essay-bucket",
		CDNDomain: "https://cdn.example.com",
	}
	c := ossutil.NewClient(cfg)

	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"downloads/a.pdf", "downloads/a.pdf", true},
		{"https://essay-bucket.oss-cn-hangzhou.aliyuncs.com/downloads/a.pdf", "downloads/a.pdf", true},
		{"https://cdn.example.com/downloads/a.pdf", "downloads/a.pdf", true},
		{"https://oss.example.com/other.pdf", "", false},
	}
	for _, tc := range cases {
		got, ok := c.ObjectKeyFromURL(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("ObjectKeyFromURL(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSignDownloadURLFallbackWhenUnconfigured(t *testing.T) {
	c := ossutil.NewClient(config.OSSConfig{})
	url := c.SignDownloadURL("https://oss.example.com/test.pdf", 300)
	if !strings.Contains(url, "expires=") {
		t.Fatalf("expected pseudo signed url, got %s", url)
	}
}

func TestPublicURLPrefersCDN(t *testing.T) {
	c := ossutil.NewClient(config.OSSConfig{
		Endpoint:  "oss-cn-hangzhou.aliyuncs.com",
		Bucket:    "essay-bucket",
		CDNDomain: "https://cdn.example.com",
	})
	if got := c.PublicURL("covers/a.jpg"); got != "https://cdn.example.com/covers/a.jpg" {
		t.Fatalf("PublicURL=%q", got)
	}
}

func TestNormalizeObjectRef(t *testing.T) {
	c := ossutil.NewClient(config.OSSConfig{
		Endpoint:  "oss-cn-hangzhou.aliyuncs.com",
		Bucket:    "essay-bucket",
		CDNDomain: "https://cdn.example.com",
	})
	if got := c.NormalizeObjectRef("https://cdn.example.com/covers/a.jpg"); got != "covers/a.jpg" {
		t.Fatalf("NormalizeObjectRef CDN URL=%q", got)
	}
	if got := c.NormalizeObjectRef("essay-mvp/covers/a.jpg"); got != "essay-mvp/covers/a.jpg" {
		t.Fatalf("NormalizeObjectRef key=%q", got)
	}
	if ext := "https://other.example/x.jpg"; c.NormalizeObjectRef(ext) != ext {
		t.Fatalf("NormalizeObjectRef external should be unchanged")
	}
}

func TestResolvePublicURL(t *testing.T) {
	c := ossutil.NewClient(config.OSSConfig{
		Endpoint:  "oss-cn-hangzhou.aliyuncs.com",
		Bucket:    "essay-bucket",
		CDNDomain: "https://cdn.example.com",
	})
	if got := c.ResolvePublicURL("covers/a.jpg"); got != "https://cdn.example.com/covers/a.jpg" {
		t.Fatalf("ResolvePublicURL key=%q", got)
	}
	cNoCDN := ossutil.NewClient(config.OSSConfig{
		Endpoint: "oss-cn-hangzhou.aliyuncs.com",
		Bucket:   "essay-bucket",
	})
	if got := cNoCDN.ResolvePublicURL("covers/a.jpg"); got != "https://essay-bucket.oss-cn-hangzhou.aliyuncs.com/covers/a.jpg" {
		t.Fatalf("ResolvePublicURL no CDN=%q", got)
	}
}
