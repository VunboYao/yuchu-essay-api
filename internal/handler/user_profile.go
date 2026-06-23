// 小程序用户资料：读取与更新昵称/头像（静默登录后可选完善）。
package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type updateUserProfileRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// GetUserProfile 返回当前登录用户资料。
func (s *Server) GetUserProfile(c *gin.Context) {
	userID := currentAuthID(c)
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "user not found")
			return
		}
		response.ServerError(c, "failed to load user profile")
		return
	}
	response.Success(c, gin.H{
		"id":       user.ID,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	})
}

// UpdateUserProfile 更新昵称或头像；至少传一项非空字段。
func (s *Server) UpdateUserProfile(c *gin.Context) {
	userID := currentAuthID(c)
	var req updateUserProfileRequest
	if !bindJSON(c, &req) {
		return
	}
	nickname := strings.TrimSpace(req.Nickname)
	avatar := strings.TrimSpace(req.Avatar)
	if nickname == "" && avatar == "" {
		response.BadRequest(c, "nickname or avatar is required")
		return
	}
	if nickname != "" && len([]rune(nickname)) > 80 {
		response.BadRequest(c, "nickname too long")
		return
	}
	if avatar != "" && len(avatar) > 500 {
		response.BadRequest(c, "avatar url too long")
		return
	}

	updates := map[string]interface{}{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}

	var user model.User
	if err := database.DB.Model(&user).Where("id = ?", userID).Updates(updates).Error; err != nil {
		response.ServerError(c, "failed to update user profile")
		return
	}
	if err := database.DB.First(&user, userID).Error; err != nil {
		response.ServerError(c, "failed to load user profile")
		return
	}
	response.Success(c, gin.H{
		"id":       user.ID,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	})
}

const maxAvatarUploadBytes = 2 << 20 // 2MB

// UploadUserAvatar 接收小程序 chooseAvatar 临时文件并持久化，返回可访问 URL。
// 一期为本地上传占位；真机 HTTP 展示限制及 OSS 改造见 docs/oss-upload-backlog.md。
func (s *Server) UploadUserAvatar(c *gin.Context) {
	userID := currentAuthID(c)
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	if file.Size <= 0 || file.Size > maxAvatarUploadBytes {
		response.BadRequest(c, "invalid file size")
		return
	}
	src, err := file.Open()
	if err != nil {
		response.ServerError(c, "failed to open upload file")
		return
	}
	defer src.Close()

	header := make([]byte, 512)
	n, readErr := src.Read(header)
	if readErr != nil && readErr != io.EOF {
		response.ServerError(c, "failed to read upload file")
		return
	}
	contentType := http.DetectContentType(header[:n])
	ext := avatarExtFromContentType(contentType)
	if ext == "" {
		response.BadRequest(c, "only jpeg or png is allowed")
		return
	}

	dir := filepath.Join("uploads", "avatars")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		response.ServerError(c, "failed to prepare upload dir")
		return
	}
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	dstPath := filepath.Join(dir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		response.ServerError(c, "failed to save upload file")
		return
	}
	defer dst.Close()
	if n > 0 {
		if _, err := dst.Write(header[:n]); err != nil {
			response.ServerError(c, "failed to save upload file")
			return
		}
	}
	if _, err := io.Copy(dst, src); err != nil {
		response.ServerError(c, "failed to save upload file")
		return
	}

	urlPath := "/uploads/avatars/" + filename
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	avatarURL := fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, urlPath)
	response.Success(c, gin.H{"url": avatarURL, "path": urlPath})
}

func avatarExtFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	default:
		return ""
	}
}
