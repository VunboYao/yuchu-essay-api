// Day9：小程序首页、分类树、已发布文章/优秀作文、搜索（仅 status=published/enabled）。
package handler

import (
	"errors"
	"strings"

	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Home 聚合推荐文章、热门文章、优秀作文与启用中的首页区块配置。
// 可选 query grade：按年级筛后再各取前 10 条；不传则仍为全局前 10。不套 articleFilters，避免学期/单元/关键词污染聚合。
func (s *Server) Home(c *gin.Context) {
	grade := strings.TrimSpace(c.Query("grade"))
	withGrade := func(db *gorm.DB) *gorm.DB {
		if grade == "" {
			return db
		}
		return db.Where("grade = ?", grade)
	}

	var recommended []model.Article
	// 推荐：已发布 + is_recommended，sort_order DESC, id DESC
	withGrade(database.DB.Where("status = ? AND is_recommended = ?", model.ArticleStatusPublished, true)).
		Omit("body_images").Order("sort_order DESC, id DESC").Limit(10).Find(&recommended)

	var hot []model.Article
	// 热门：已发布 + is_hot，view_count DESC, id DESC
	withGrade(database.DB.Where("status = ? AND is_hot = ?", model.ArticleStatusPublished, true)).
		Omit("body_images").Order("view_count DESC, id DESC").Limit(10).Find(&hot)

	var essays []model.EssayExample
	// 优秀作文：已启用，sort_order DESC, id DESC
	withGrade(database.DB.Where("status = ?", model.StatusEnabled)).
		Order("sort_order DESC, id DESC").Limit(10).Find(&essays)

	var blocks []model.HomeBlock
	// 查询首页区块
	database.DB.Where("status = ?", model.StatusEnabled).
		Order("sort_order DESC, id DESC").Find(&blocks)

	s.resolveArticlesMedia(recommended)
	s.resolveArticlesMedia(hot)
	s.resolveHomeBlocksMedia(blocks)

	response.Success(c, gin.H{
		"recommendedArticles": recommended,
		"hotArticles":         hot,
		"essayExamples":       essays,
		"homeBlocks":          blocks,
	})
}

// PublicAdSlots 返回启用中的广告位配置，供小程序按 slotKey 渲染流量主组件。
// 可选 query page=home|article_detail|essay_detail|search|mine 按所属页面过滤。
func (s *Server) PublicAdSlots(c *gin.Context) {
	q := database.DB.Where("status = ?", model.StatusEnabled)
	if page := strings.TrimSpace(c.Query("page")); page != "" {
		q = q.Where("page = ?", page)
	}
	var list []model.AdSlot
	if err := q.Order("sort_order DESC, id DESC").Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list ad slots")
		return
	}
	response.Success(c, list)
}

// PublicHotSearchKeywords 返回启用中的搜索热门词（仅 keyword 字符串，按 sort_order DESC）。
func (s *Server) PublicHotSearchKeywords(c *gin.Context) {
	var list []model.HotSearchKeyword
	if err := database.DB.Where("status = ?", model.StatusEnabled).
		Order("sort_order DESC, id DESC").
		Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list hot search keywords")
		return
	}
	keywords := make([]string, 0, len(list))
	for _, item := range list {
		keywords = append(keywords, item.Keyword)
	}
	response.Success(c, keywords)
}

// PublicOfficialAccountConfig 返回指定 scene 下启用中的公众号导流配置，供小程序展示关注引导。
func (s *Server) PublicOfficialAccountConfig(c *gin.Context) {
	scene := strings.TrimSpace(c.Query("scene"))
	if scene == "" {
		response.BadRequest(c, "scene is required")
		return
	}
	cfg, err := model.FindOfficialAccountConfigByScene(database.DB, scene)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Success(c, nil)
			return
		}
		response.ServerError(c, "failed to load official account config")
		return
	}
	if cfg.Status != model.StatusEnabled {
		response.Success(c, nil)
		return
	}
	if strings.TrimSpace(cfg.Username) == "" {
		response.Success(c, nil)
		return
	}
	response.Success(c, gin.H{
		"scene":       cfg.Scene,
		"title":       cfg.Title,
		"description": cfg.Description,
		"username":    cfg.Username,
	})
}

// Categories 查询教材分类单元
func (s *Server) Categories(c *gin.Context) {
	var list []model.TaxonomyUnit
	if err := database.DB.Where("status = ?", model.StatusEnabled).
		Order("grade ASC, semester ASC, unit ASC, sort_order DESC").Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list categories")
		return
	}
	response.Success(c, list)
}

// PublicArticles 查询已发布文章
func (s *Server) PublicArticles(c *gin.Context) {
	page, pageSize := pageParams(c)
	// 构建查询条件
	db := articleFilters(database.DB.Model(&model.Article{}), c).Where("status = ?", model.ArticleStatusPublished)
	var total int64
	db.Count(&total)
	var list []model.Article
	if err := paginate(db.Omit("body_images").Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list articles")
		return
	}
	s.resolveArticlesMedia(list)
	response.Page(c, list, total, page, pageSize)
}

// PublicArticleDetail 已发布文章详情；递增 view_count，并返回该文章下可免费下载的资料列表。
func (s *Server) PublicArticleDetail(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.Article
	if err := database.DB.Preload("Essays", "status = ?", model.StatusEnabled).
		Where("status = ?", model.ArticleStatusPublished).First(&item, id).Error; err != nil {
		response.NotFound(c, "article not found")
		return
	}
	database.DB.Model(&item).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	// 查询可免费下载的资料列表
	var downloads []model.DownloadFile
	database.DB.Where("owner_type = ? AND owner_id = ? AND status = ? AND access_type = ? AND tags LIKE ?",
		"article", item.ID, model.StatusEnabled, model.AccessTypeFree, "%免费%").
		Order("sort_order DESC, id DESC").Find(&downloads)

	relatedArticles := listRelatedArticles(item, relatedArticleLimit)

	s.resolveArticleMedia(&item)
	s.resolveArticlesMedia(relatedArticles)

	response.Success(c, gin.H{
		"article":         item,
		"relatedArticles": relatedArticles,
		"freeDownloads":   downloads,
		"downloadGate":    buildDownloadGateResponse(),
	})
}

const relatedArticleLimit = 10

// RelatedArticleMatches 相关推荐须与当前文章同年级、同册次、同单元。
func RelatedArticleMatches(baseGrade int, baseSemester string, baseUnit int, grade int, semester string, unit int) bool {
	return grade == baseGrade && semester == baseSemester && unit == baseUnit
}

// listRelatedArticles 返回同年级/同册次/同单元的其他已发布文章，排除当前篇。
func listRelatedArticles(article model.Article, limit int) []model.Article {
	if limit <= 0 {
		limit = relatedArticleLimit
	}
	var list []model.Article
	database.DB.Model(&model.Article{}).
		Where("status = ? AND id <> ? AND grade = ? AND semester = ? AND unit = ?",
			model.ArticleStatusPublished, article.ID, article.Grade, article.Semester, article.Unit).
		Omit("body_images").
		Order("sort_order DESC, id DESC").
		Limit(limit).
		Find(&list)
	return list
}

// buildDownloadGateResponse 聚合全局下载激励视频门槛，供文章详情资料区使用。
func buildDownloadGateResponse() model.DownloadGateInfo {
	gate, err := model.FindDownloadGateSlot(database.DB)
	if err != nil {
		return model.DownloadGateInfo{Required: false}
	}
	return model.BuildDownloadGate(gate)
}

func (s *Server) PublicEssays(c *gin.Context) {
	page, pageSize := pageParams(c)
	db := essayFilters(database.DB.Model(&model.EssayExample{}), c).
		Where("status = ?", model.StatusEnabled)
	var total int64
	db.Count(&total)
	var list []model.EssayExample
	if err := paginate(db.Order("sort_order DESC, id DESC"), page, pageSize).Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list essays")
		return
	}
	essayIDs := make([]uint, 0, len(list))
	for _, item := range list {
		essayIDs = append(essayIDs, item.ID)
	}
	items := BuildEssayListItems(list, essayTagsByEssayIDs(essayIDs))
	payload := gin.H{
		"list":     items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}
	// 首屏附带可选标签，供小程序筛选栏渲染；total 已含当前全部筛选条件
	if page == 1 {
		payload["filterTags"] = listPublicEssayFilterTags(c)
	}
	response.Success(c, payload)
}

func (s *Server) PublicEssayDetail(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var item model.EssayExample
	if err := database.DB.Where("status = ?", model.StatusEnabled).First(&item, id).Error; err != nil {
		response.NotFound(c, "essay not found")
		return
	}
	database.DB.Model(&item).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	resp := gin.H{"essay": item, "article": nil}
	if item.ArticleID != nil && *item.ArticleID > 0 {
		var article model.Article
		if err := database.DB.Select("id", "title", "grade", "semester", "unit", "essay_topic").
			Where("status = ?", model.ArticleStatusPublished).
			First(&article, *item.ArticleID).Error; err == nil {
			resp["article"] = gin.H{
				"id":         article.ID,
				"title":      article.Title,
				"grade":      article.Grade,
				"semester":   article.Semester,
				"unit":       article.Unit,
				"essayTopic": article.EssayTopic,
			}
		}
	}
	response.Success(c, resp)
}

// searchResultLimitPerType 搜索每种内容类型最多返回条数（不分页）。
const searchResultLimitPerType = 20

// SearchResultItem 搜索混排列表项；targetType 区分文章/优秀作文，供小程序选卡片与跳转。
type SearchResultItem struct {
	TargetType    string `json:"targetType"`
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	CoverImg      string `json:"coverImg,omitempty"`
	Grade         int    `json:"grade"`
	Semester      string `json:"semester,omitempty"`
	Unit          int    `json:"unit,omitempty"`
	Summary       string `json:"summary,omitempty"`
	IsRecommended bool   `json:"isRecommended,omitempty"`
	IsHot         bool   `json:"isHot,omitempty"`
	WordCount     int    `json:"wordCount,omitempty"`
}

// Search 检索已发布文章与启用中的优秀作文；每类最多 20 条，先文章后作文；埋点仍走 POST /track/search。
func (s *Server) Search(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		response.BadRequest(c, "keyword is required")
		return
	}
	like := "%" + keyword + "%"

	// 文章：标题 / 作文题目 / 摘要
	var articles []model.Article
	if err := articleFilters(database.DB.Model(&model.Article{}), c).
		Where("status = ?", model.ArticleStatusPublished).
		Where("title LIKE ? OR essay_topic LIKE ? OR summary LIKE ?", like, like, like).
		Omit("body_images").
		Order("view_count DESC, id DESC").
		Limit(searchResultLimitPerType).
		Find(&articles).Error; err != nil {
		response.ServerError(c, "failed to search")
		return
	}
	s.resolveArticlesMedia(articles)

	// 作文：标题 / 正文匹配，但不 Select body，避免把 longtext 打进列表
	var essays []model.EssayExample
	if err := database.DB.Model(&model.EssayExample{}).
		Select("id", "article_id", "title", "grade", "word_count", "status", "sort_order", "view_count", "favorite_count", "created_at", "updated_at").
		Where("status = ?", model.StatusEnabled).
		Where("title LIKE ? OR body LIKE ?", like, like).
		Order("view_count DESC, id DESC").
		Limit(searchResultLimitPerType).
		Find(&essays).Error; err != nil {
		response.ServerError(c, "failed to search")
		return
	}

	list := BuildSearchResultList(articles, essays)
	total := int64(len(list))
	response.Page(c, list, total, 1, int(total))
}

// BuildSearchResultList 先文章后作文，映射为统一搜索卡片字段。
func BuildSearchResultList(articles []model.Article, essays []model.EssayExample) []SearchResultItem {
	list := make([]SearchResultItem, 0, len(articles)+len(essays))
	for _, a := range articles {
		list = append(list, SearchResultItem{
			TargetType:    model.TargetTypeArticle,
			ID:            a.ID,
			Title:         a.Title,
			CoverImg:      a.CoverImg,
			Grade:         a.Grade,
			Semester:      a.Semester,
			Unit:          a.Unit,
			Summary:       a.Summary,
			IsRecommended: a.IsRecommended,
			IsHot:         a.IsHot,
		})
	}
	for _, e := range essays {
		list = append(list, SearchResultItem{
			TargetType: model.TargetTypeEssay,
			ID:         e.ID,
			Title:      e.Title,
			Grade:      e.Grade,
			WordCount:  e.WordCount,
		})
	}
	return list
}
