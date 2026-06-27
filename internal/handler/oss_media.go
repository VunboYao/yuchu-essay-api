// OSS 媒体字段读写：库内 objectKey，读时拼 CDN / OSS 直链。
package handler

import "course_server/internal/model"

// MediaURLResolver 将库内 OSS 引用解析为对外 HTTPS URL。
type MediaURLResolver func(string) string

func (s *Server) mediaResolver() MediaURLResolver {
	return s.OSS.ResolvePublicURL
}

func (s *Server) applyArticleWriteMedia(a *model.Article) {
	if a == nil {
		return
	}
	a.CoverImg = s.OSS.NormalizeObjectRef(a.CoverImg)
	a.BodyImages = s.OSS.NormalizeObjectRefs(a.BodyImages)
}

func (s *Server) resolveArticleMedia(a *model.Article) {
	if a == nil {
		return
	}
	a.CoverImg = s.OSS.ResolvePublicURL(a.CoverImg)
	for i, u := range a.BodyImages {
		a.BodyImages[i] = s.OSS.ResolvePublicURL(u)
	}
}

func (s *Server) resolveArticlesMedia(list []model.Article) {
	for i := range list {
		s.resolveArticleMedia(&list[i])
	}
}

func (s *Server) applyHomeBlockWriteMedia(b *model.HomeBlock) {
	if b == nil {
		return
	}
	b.Icon = s.OSS.NormalizeObjectRef(b.Icon)
}

func (s *Server) resolveHomeBlockMedia(b *model.HomeBlock) {
	if b == nil {
		return
	}
	b.Icon = s.OSS.ResolvePublicURL(b.Icon)
}

func (s *Server) resolveHomeBlocksMedia(list []model.HomeBlock) {
	for i := range list {
		s.resolveHomeBlockMedia(&list[i])
	}
}

func (s *Server) applyDownloadFileWriteMedia(f *model.DownloadFile) {
	if f == nil {
		return
	}
	f.FileURL = s.OSS.NormalizeObjectRef(f.FileURL)
}

func resolveContentArticleTarget(target any, resolve MediaURLResolver) any {
	if resolve == nil {
		return target
	}
	summary, ok := target.(ContentArticleSummary)
	if !ok {
		return target
	}
	summary.CoverImg = resolve(summary.CoverImg)
	return summary
}
