// 后台运营配置：首页区块、广告位、公众号引导；OSS 上传为占位（Day7）。
package handler

import (
	"errors"
	"strings"

	"course_server/internal/database"
	admindto "course_server/internal/dto/admin"
	"course_server/internal/middleware"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListHomeBlocks 查询首页入口列表
func (s *Server) ListHomeBlocks(c *gin.Context) {
	var list []model.HomeBlock
	if err := database.DB.Order("sort_order DESC, id DESC").Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list home blocks")
		return
	}
	response.Success(c, list)
}

// CreateHomeBlock 创建首页入口
func (s *Server) CreateHomeBlock(c *gin.Context) {
	var req admindto.CreateHomeBlockRequest
	if !bindJSON(c, &req) {
		return
	}
	block, err := req.ValidateForCreate()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if block.SortOrder == 0 {
		block.SortOrder = model.NextSortOrder(database.DB, &model.HomeBlock{})
	}
	if err := database.DB.Create(&block).Error; err != nil {
		response.ServerError(c, "failed to create home block")
		return
	}
	middleware.SetOpLogRemark(c, "创建首页入口："+block.Title)
	middleware.SetOpLogTargetID(c, block.ID)
	response.Success(c, block)
}

// UpdateHomeBlock 更新首页入口
func (s *Server) UpdateHomeBlock(c *gin.Context) {
	// 获取路径参数 ID
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	// 按 ID 查找已有记录
	existing, findErr := model.FindHomeBlockByID(database.DB, uint(id))
	if findErr != nil {
		// 如果记录不存在，则返回 404
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			response.NotFound(c, "home block not found")
			return
		}
		response.ServerError(c, "failed to update home block")
		return
	}
	// 绑定请求参数
	var req admindto.UpdateHomeBlockRequest
	if !bindJSON(c, &req) {
		return
	}
	// 校验请求参数
	block, err := req.ValidateForUpdate(existing)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 更新记录
	if err := UpdateAndReload(database.DB, &existing, block.UpdateMap()); err != nil {
		response.ServerError(c, "failed to update home block")
		return
	}
	// 设置操作日志
	middleware.SetOpLogRemark(c, "更新首页入口："+existing.Title)
	// 返回成功
	response.Success(c, existing)
}

// DeleteHomeBlock 物理删除首页入口；须已禁用。
func (s *Server) DeleteHomeBlock(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	item, findErr := model.FindHomeBlockByID(database.DB, uint(id))
	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			response.NotFound(c, "home block not found")
			return
		}
		response.ServerError(c, "failed to find home block")
		return
	}
	if err := model.CheckHomeBlockDeletable(item); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := database.DB.Delete(&item).Error; err != nil {
		response.ServerError(c, "failed to delete home block")
		return
	}
	middleware.SetOpLogRemark(c, "删除首页入口："+item.Title)
	response.Success(c, gin.H{"id": id})
}

func (s *Server) ListAdSlots(c *gin.Context) {
	var list []model.AdSlot
	if err := database.DB.Order("sort_order DESC, id DESC").Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list ad slots")
		return
	}
	response.Success(c, list)
}

// UpsertAdSlot 创建或更新广告位配置
func (s *Server) UpsertAdSlot(c *gin.Context) {
	var req admindto.UpsertAdSlotRequest
	if !bindJSON(c, &req) {
		return
	}
	existing, findErr := model.FindAdSlotByKey(database.DB, strings.TrimSpace(req.SlotKey))
	// 如果记录不存在，则返回 server error
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		response.ServerError(c, "failed to find ad slot")
		return
	}
	defaultStatus := model.StatusEnabled
	if findErr == nil {
		defaultStatus = existing.Status
	}
	slot, err := req.Validate(defaultStatus)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 更新时未传 adType 则保留原值，避免误覆盖为 banner
	if findErr == nil && strings.TrimSpace(req.AdType) == "" {
		slot.AdType = existing.AdType
	}
	if errors.Is(findErr, gorm.ErrRecordNotFound) {
		if err := database.DB.Create(&slot).Error; err != nil {
			response.ServerError(c, "failed to create ad slot")
			return
		}
		middleware.SetOpLogRemark(c, "创建广告位："+slot.SlotKey)
		middleware.SetOpLogTargetID(c, slot.ID)
		response.Success(c, slot)
		return
	}
	if err := UpdateAndReload(database.DB, &existing, slot.UpdateMap()); err != nil {
		response.ServerError(c, "failed to update ad slot")
		return
	}
	middleware.SetOpLogRemark(c, "更新广告位："+existing.SlotKey)
	response.Success(c, existing)
}

// DeleteAdSlot 物理删除广告位；须已禁用。
func (s *Server) DeleteAdSlot(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	item, findErr := model.FindAdSlotByID(database.DB, uint(id))
	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			response.NotFound(c, "ad slot not found")
			return
		}
		response.ServerError(c, "failed to find ad slot")
		return
	}
	if err := model.CheckAdSlotDeletable(item); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := database.DB.Delete(&item).Error; err != nil {
		response.ServerError(c, "failed to delete ad slot")
		return
	}
	middleware.SetOpLogRemark(c, "删除广告位："+item.SlotKey)
	response.Success(c, gin.H{"id": id})
}

func (s *Server) ListOfficialAccountConfigs(c *gin.Context) {
	var list []model.OfficialAccountConfig
	if err := database.DB.Order("id DESC").Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list official account configs")
		return
	}
	response.Success(c, list)
}

func (s *Server) UpsertOfficialAccountConfig(c *gin.Context) {
	var req admindto.UpsertOfficialAccountConfigRequest
	if !bindJSON(c, &req) {
		return
	}
	existing, findErr := model.FindOfficialAccountConfigByScene(database.DB, strings.TrimSpace(req.Scene))
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		response.ServerError(c, "failed to find official account config")
		return
	}
	defaultStatus := model.StatusEnabled
	if findErr == nil {
		defaultStatus = existing.Status
	}
	cfg, err := req.Validate(defaultStatus)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(findErr, gorm.ErrRecordNotFound) {
		if err := database.DB.Create(&cfg).Error; err != nil {
			response.ServerError(c, "failed to create official account config")
			return
		}
		middleware.SetOpLogRemark(c, "创建公众号配置："+cfg.Scene)
		middleware.SetOpLogTargetID(c, cfg.ID)
		response.Success(c, cfg)
		return
	}
	if err := UpdateAndReload(database.DB, &existing, cfg.UpdateMap()); err != nil {
		response.ServerError(c, "failed to update official account config")
		return
	}
	middleware.SetOpLogRemark(c, "更新公众号配置："+existing.Scene)
	response.Success(c, existing)
}

// DeleteOfficialAccountConfig 物理删除公众号导流配置；须已禁用。
func (s *Server) DeleteOfficialAccountConfig(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	item, findErr := model.FindOfficialAccountConfigByID(database.DB, uint(id))
	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			response.NotFound(c, "official account config not found")
			return
		}
		response.ServerError(c, "failed to find official account config")
		return
	}
	if err := model.CheckOfficialAccountConfigDeletable(item); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := database.DB.Delete(&item).Error; err != nil {
		response.ServerError(c, "failed to delete official account config")
		return
	}
	middleware.SetOpLogRemark(c, "删除公众号配置："+item.Scene)
	response.Success(c, gin.H{"id": id})
}

// UploadSTS 签发 OSS STS 临时凭证供 Admin 浏览器直传；未配置时返回占位。
func (s *Server) UploadSTS(c *gin.Context) {
	middleware.SetOpLogRemark(c, "获取 OSS 上传凭证")
	if !s.OSS.IsSTSEnabled() {
		response.Success(c, gin.H{
			"mode":    "placeholder",
			"message": "configure oss endpoint/bucket/access_key and sts_role_arn before enabling direct upload",
		})
		return
	}
	cred, err := s.OSS.IssueSTSCredentials()
	if err != nil {
		response.ServerError(c, "failed to issue oss sts credentials")
		return
	}
	response.Success(c, cred)
}

// DeleteByID 通用删除 stub，未在 router 注册，勿调用。
func (s *Server) DeleteByID(modelName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := pathID(c)
		if err != nil {
			response.BadRequest(c, "invalid id")
			return
		}
		response.Success(c, gin.H{"model": modelName, "id": id})
	}
}
