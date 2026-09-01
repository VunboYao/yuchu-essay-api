package htmlsanitize_test

import (
	"strings"
	"testing"

	"course_server/internal/pkg/htmlsanitize"
)

func TestSanitizeStripsScript(t *testing.T) {
	t.Parallel()
	in := `<p>正文</p><script>alert(1)</script>`
	out := htmlsanitize.Sanitize(in)
	if strings.Contains(out, "script") {
		t.Fatalf("expected script removed, got %q", out)
	}
	if !strings.Contains(out, "正文") {
		t.Fatalf("expected body text kept, got %q", out)
	}
}

func TestSanitizeKeepsFormatting(t *testing.T) {
	t.Parallel()
	in := `<p style="text-indent:2em">段一</p><p><strong>加粗</strong><u>下划线</u></p>`
	out := htmlsanitize.Sanitize(in)
	if !strings.Contains(out, "text-indent") {
		t.Fatalf("expected text-indent kept, got %q", out)
	}
	if !strings.Contains(out, "<strong>") || !strings.Contains(out, "<u>") {
		t.Fatalf("expected formatting tags kept, got %q", out)
	}
}

func TestSanitizeAddsParagraphIndent(t *testing.T) {
	t.Parallel()
	out := htmlsanitize.Sanitize(`<p>段一</p>`)
	if !strings.Contains(out, `text-indent:2em`) {
		t.Fatalf("expected default indent injected, got %q", out)
	}
}

func TestSanitizeKeepsTextAlign(t *testing.T) {
	t.Parallel()
	in := `<p style="text-align:center">标题</p>`
	out := htmlsanitize.Sanitize(in)
	if !strings.Contains(out, "text-align") {
		t.Fatalf("expected text-align kept, got %q", out)
	}
	if strings.Contains(out, "text-indent") {
		t.Fatalf("expected no text-indent for centered paragraph, got %q", out)
	}
}

func TestSanitizeAddsIndentForLeftAlignedParagraph(t *testing.T) {
	t.Parallel()
	out := htmlsanitize.Sanitize(`<p style="text-align:left">段一</p>`)
	if !strings.Contains(out, "text-align") {
		t.Fatalf("expected text-align kept, got %q", out)
	}
	if !strings.Contains(out, `text-indent:2em`) {
		t.Fatalf("expected indent for left-aligned paragraph, got %q", out)
	}
}

func TestSanitizeKeepsZeroIndent(t *testing.T) {
	t.Parallel()
	for _, in := range []string{
		`<p style="text-indent:0">亲爱的老师：</p><p>正文</p>`,
		`<p style="text-indent:0em">亲爱的老师：</p><p>正文</p>`,
	} {
		out := htmlsanitize.Sanitize(in)
		idx := strings.Index(out, "亲爱的老师")
		if idx < 0 {
			t.Fatalf("expected greeting kept, got %q", out)
		}
		open := strings.LastIndex(out[:idx], "<p")
		gt := strings.Index(out[open:], ">")
		if open < 0 || gt < 0 {
			t.Fatalf("expected opening p tag before greeting, got %q", out)
		}
		tag := out[open : open+gt+1]
		if strings.Contains(tag, "2em") {
			t.Fatalf("greeting paragraph should keep zero indent, tag=%q html=%q", tag, out)
		}
		if !strings.Contains(strings.ToLower(tag), "text-indent") {
			t.Fatalf("greeting paragraph should keep text-indent:0, tag=%q html=%q", tag, out)
		}
		if !strings.Contains(out, `text-indent:2em`) {
			t.Fatalf("expected following paragraph still indented, got %q", out)
		}
	}
}

func TestStripTags(t *testing.T) {
	t.Parallel()
	plain := htmlsanitize.StripTags(`<p>你好 <strong>世界</strong></p>`)
	if plain != "你好 世界" {
		t.Fatalf("expected 你好 世界, got %q", plain)
	}
}

func TestIsEffectivelyEmpty(t *testing.T) {
	t.Parallel()
	cases := []struct {
		html  string
		empty bool
	}{
		{"<p><br></p>", true},
		{"<p> </p>", true},
		{"<p>有内容</p>", false},
	}
	for _, c := range cases {
		if got := htmlsanitize.IsEffectivelyEmpty(c.html); got != c.empty {
			t.Fatalf("IsEffectivelyEmpty(%q) = %v, want %v", c.html, got, c.empty)
		}
	}
}
