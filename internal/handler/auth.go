// Day3/Day8：管理员与小程序登录鉴权。
package handler

import (
	"errors"
	"time"

	"course_server/internal/database"
	"course_server/internal/model"
	jwtutil "course_server/internal/pkg/jwt"
	"course_server/internal/pkg/response"
	wxutil "course_server/internal/pkg/wechat"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type adminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLogin 后台登录：校验启用状态管理员 + bcrypt 密码，签发 admin JWT。
func (s *Server) AdminLogin(c *gin.Context) {
	var req adminLoginRequest
	// 绑定请求参数
	if !bindJSON(c, &req) {
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
// 未配置 wechat 凭证时 openid = "dev_" + code（本地联调）；配置后走 code2session。
func (s *Server) WxLogin(c *gin.Context) {
	// 1. 绑定请求参数
	var req wxLoginRequest
	if !bindJSON(c, &req) {
		return
	}
	openid := "dev_" + req.Code
	unionid := ""
	if !wxutil.DevMode(s.Config.Wechat.AppID, s.Config.Wechat.AppSecret) {
		sess, err := wxutil.Code2Session(s.Config.Wechat.AppID, s.Config.Wechat.AppSecret, req.Code)
		if err != nil {
			response.BadRequest(c, "wechat login failed")
			return
		}
		openid = sess.OpenID
		unionid = sess.UnionID
	}
	// 3. 查询用户
	var user model.User
	err := database.DB.Where("open_id = ?", openid).First(&user).Error
	// 仅「记录不存在」才创建；其他 DB 错误直接返回，避免把瞬时故障误判为新用户
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.ServerError(c, "failed to query user")
		return
	}
	if err != nil {
		// 4. 创建用户
		user = model.User{
			OpenID:      openid,
			UnionID:     unionid,
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
		// 5. 更新用户信息
		updates := map[string]interface{}{}
		if unionid != "" && user.UnionID != unionid {
			updates["union_id"] = unionid
		}
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

	// 6. 签发用户 JWT
	token, err := jwtutil.SignUser(s.Config.JWT.Secret, s.Config.JWT.Issuer, user.ID, user.OpenID)
	if err != nil {
		response.ServerError(c, "failed to sign token")
		return
	}
	// 7. 返回响应
	response.Success(c, gin.H{"token": token, "user": user})
}
