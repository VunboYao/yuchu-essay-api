// Day3/Day8：管理员与小程序登录鉴权。
package handler

import (
	"time"

	"course_server/internal/database"
	"course_server/internal/model"
	jwtutil "course_server/internal/pkg/jwt"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type adminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLogin 后台登录：校验启用状态管理员 + bcrypt 密码，签发 admin JWT。
func (s *Server) AdminLogin(c *gin.Context) {
	var req adminLoginRequest
	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var admin model.Admin
	// 查询启用状态的管理员
	if err := database.DB.Where("username = ? AND status = ?", req.Username, model.StatusEnabled).First(&admin).Error; err != nil {
		response.Unauthorized(c, "invalid username or password")
		return
	}
	// 比较密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		response.Unauthorized(c, "invalid username or password")
		return
	}

	// 更新最近登录时间
	now := time.Now()
	// 更新最近登录时间
	database.DB.Model(&admin).Update("last_login_at", &now)

	// 签发管理员 JWT
	token, err := jwtutil.SignAdmin(s.Config.JWT.Secret, s.Config.JWT.Issuer, admin.ID, admin.Username, admin.Role)
	if err != nil {
		response.ServerError(c, "failed to sign token")
		return
	}
	response.Success(c, gin.H{
		"token":    token,
		"adminId":  admin.ID,
		"username": admin.Username,
		"role":     admin.Role,
	})
}

type wxLoginRequest struct {
	Code        string `json:"code" binding:"required"`
	SourceScene string `json:"sourceScene"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
}

// WxLogin 小程序登录（Day8）：按 openid 查找或创建用户并签发 user JWT。
// 当前为开发占位：openid = "dev_" + code；配置 wechat 后应改为 code2session。
func (s *Server) WxLogin(c *gin.Context) {
	var req wxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	openid := "dev_" + req.Code
	var user model.User
	err := database.DB.Where("open_id = ?", openid).First(&user).Error
	if err != nil {
		user = model.User{
			OpenID:      openid,
			SourceScene: req.SourceScene,
			Nickname:    req.Nickname,
			Avatar:      req.Avatar,
			Status:      model.StatusEnabled,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			response.ServerError(c, "failed to create user")
			return
		}
	} else {
		updates := map[string]interface{}{}
		if req.SourceScene != "" {
			updates["source_scene"] = req.SourceScene
		}
		if req.Nickname != "" {
			updates["nickname"] = req.Nickname
		}
		if req.Avatar != "" {
			updates["avatar"] = req.Avatar
		}
		if len(updates) > 0 {
			database.DB.Model(&user).Updates(updates)
		}
	}

	token, err := jwtutil.SignUser(s.Config.JWT.Secret, s.Config.JWT.Issuer, user.ID, user.OpenID)
	if err != nil {
		response.ServerError(c, "failed to sign token")
		return
	}
	response.Success(c, gin.H{"token": token, "user": user})
}
