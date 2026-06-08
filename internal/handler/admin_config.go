// 后台运营配置：首页区块、广告位、公众号引导；OSS 上传为占位（Day7）。
package handler

import (
	"errors"
	"strings"

	"course_server/internal/database"
	admindto "course_server/internal/dto/admin"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (s *Server) ListHomeBlocks(c *gin.Context) {
	var list []model.HomeBlock
	if err := database.DB.Order("sort_order DESC, id DESC").Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list home blocks")
		return
	}
	response.Success(c, list)
}

func (s *Server) UpsertHomeBlock(c *gin.Context) {
	var req admindto.UpsertHomeBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	existing, findErr := model.FindHomeBlockByKey(database.DB, strings.TrimSpace(req.BlockKey))
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		response.ServerError(c, "failed to save home block")
		return
	}
	defaultStatus := model.StatusEnabled
	if findErr == nil {
		defaultStatus = existing.Status
	}
	block, err := req.Validate(defaultStatus)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(findErr, gorm.ErrRecordNotFound) {
		if err := database.DB.Create(&block).Error; err != nil {
			response.ServerError(c, "failed to save home block")
			return
		}
		response.Success(c, block)
		return
	}
	if err := UpdateAndReload(database.DB, &existing, block.UpdateMap()); err != nil {
		response.ServerError(c, "failed to save home block")
		return
	}
	response.Success(c, existing)
}

func (s *Server) ListAdSlots(c *gin.Context) {
	var list []model.AdSlot
	if err := database.DB.Order("sort_order DESC, id DESC").Find(&list).Error; err != nil {
		response.ServerError(c, "failed to list ad slots")
		return
	}
	response.Success(c, list)
}

func (s *Server) UpsertAdSlot(c *gin.Context) {
	var req admindto.UpsertAdSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	existing, findErr := model.FindAdSlotByKey(database.DB, strings.TrimSpace(req.SlotKey))
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		response.ServerError(c, "failed to save ad slot")
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
	if errors.Is(findErr, gorm.ErrRecordNotFound) {
		if err := database.DB.Create(&slot).Error; err != nil {
			response.ServerError(c, "failed to save ad slot")
			return
		}
		response.Success(c, slot)
		return
	}
	if err := UpdateAndReload(database.DB, &existing, slot.UpdateMap()); err != nil {
		response.ServerError(c, "failed to save ad slot")
		return
	}
	response.Success(c, existing)
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
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	existing, findErr := model.FindOfficialAccountConfigByScene(database.DB, strings.TrimSpace(req.Scene))
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		response.ServerError(c, "failed to save official account config")
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
			response.ServerError(c, "failed to save official account config")
			return
		}
		response.Success(c, cfg)
		return
	}
	if err := UpdateAndReload(database.DB, &existing, cfg.UpdateMap()); err != nil {
		response.ServerError(c, "failed to save official account config")
		return
	}
	response.Success(c, existing)
}

// UploadSTS 封面上传临时凭证（Day7 占位）；需配置 config.yaml 中 oss 后实现真实 STS。
func (s *Server) UploadSTS(c *gin.Context) {
	response.Success(c, gin.H{
		"mode":    "placeholder",
		"message": "configure OSS credentials before enabling direct upload",
	})
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
