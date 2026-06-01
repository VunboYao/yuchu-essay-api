// 后台运营配置：首页区块、广告位、公众号引导；OSS 上传为占位（Day7）。
package handler

import (
	"strconv"

	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
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
	var req model.HomeBlock
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Status == 0 {
		req.Status = model.StatusEnabled
	}
	if err := database.DB.Save(&req).Error; err != nil {
		response.ServerError(c, "failed to save home block")
		return
	}
	response.Success(c, req)
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
	var req model.AdSlot
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Status == 0 {
		req.Status = model.StatusEnabled
	}
	if err := database.DB.Save(&req).Error; err != nil {
		response.ServerError(c, "failed to save ad slot")
		return
	}
	response.Success(c, req)
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
	var req model.OfficialAccountConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Status == 0 {
		req.Status = model.StatusEnabled
	}
	if err := database.DB.Save(&req).Error; err != nil {
		response.ServerError(c, "failed to save official account config")
		return
	}
	response.Success(c, req)
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
		id, _ := strconv.Atoi(c.Param("id"))
		response.Success(c, gin.H{"model": modelName, "id": id})
	}
}
