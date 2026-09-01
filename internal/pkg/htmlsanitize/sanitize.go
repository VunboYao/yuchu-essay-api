// Package htmlsanitize 优秀作文富文本白名单清洗与纯文本提取。
package htmlsanitize

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
)

var paragraphOpenTag = regexp.MustCompile(`(?i)<p(\s+style="([^"]*)")?\s*>`)

var (
	essayPolicy = newEssayPolicy()
	stripPolicy = bluemonday.StrictPolicy()
)

func newEssayPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "br", "strong", "b", "em", "i", "u", "mark", "span")
	p.AllowAttrs("style").OnElements("p", "span")
	p.AllowStyles("color", "background-color").OnElements("span")
	p.AllowStyles("text-indent", "text-align").OnElements("p")
	return p
}

// Sanitize 按优秀作文白名单清洗 HTML，剔除 script 等危险内容，并补全段落两字缩进。
func Sanitize(html string) string {
	clean := essayPolicy.Sanitize(html)
	return ensureParagraphIndent(clean)
}

// paragraphAlignSkipsIndent 居中/居右段落不再补两字缩进。
func paragraphAlignSkipsIndent(style string) bool {
	compact := strings.ReplaceAll(strings.ToLower(style), " ", "")
	return strings.Contains(compact, "text-align:center") || strings.Contains(compact, "text-align:right")
}

// ensureParagraphIndent 为 p 标签补 text-indent:2em，供小程序 rich-text 渲染标准段落缩进。
func ensureParagraphIndent(html string) string {
	return paragraphOpenTag.ReplaceAllStringFunc(html, func(tag string) string {
		parts := paragraphOpenTag.FindStringSubmatch(tag)
		if len(parts) > 2 && parts[2] != "" {
			style := parts[2]
			// 已有 text-indent（含 0 / 0em，写信称谓顶格）或居中/居右则不覆盖
			if strings.Contains(strings.ToLower(style), "text-indent") || paragraphAlignSkipsIndent(style) {
				return tag
			}
			return `<p style="` + style + `;text-indent:2em">`
		}
		return `<p style="text-indent:2em">`
	})
}

// StripTags 去除全部 HTML 标签，仅保留可见文本（用于字数统计与空内容校验）。
func StripTags(html string) string {
	return stripPolicy.Sanitize(html)
}

// IsEffectivelyEmpty 判断富文本是否无实质内容（如仅 <p><br></p>）。
func IsEffectivelyEmpty(html string) bool {
	plain := strings.TrimSpace(StripTags(html))
	if plain == "" {
		return true
	}
	for _, r := range plain {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
