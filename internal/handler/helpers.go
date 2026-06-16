// 分页与鉴权上下文辅助函数，供各 handler 复用。
package handler

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"course_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// pathID 解析路径 :id，须为正整数；非法时返回 error。
func pathID(c *gin.Context) (int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}

// pageParams 解析 page/pageSize，默认 1/20，pageSize 上限 100。
func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

// paginate 分页查询。
func paginate(db *gorm.DB, page, pageSize int) *gorm.DB {
	// 偏移量
	return db.Offset((page - 1) * pageSize).Limit(pageSize)
}

// bindJSON 绑定 JSON 请求体；失败时返回 400 与可读 message。
func bindJSON(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		response.BadRequest(c, JSONBindErrorMessage(err))
		return false
	}
	return true
}

// JSONBindErrorMessage 将绑定错误转为对外可读文案，避免直接暴露 EOF 等底层错误。
func JSONBindErrorMessage(err error) string {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return "request body is required"
	}
	var syntax *json.SyntaxError
	if errors.As(err, &syntax) {
		return "invalid JSON body"
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return jsonFieldLabel(typeErr.Field) + " has invalid type"
	}
	var ve validator.ValidationErrors
	if errors.As(err, &ve) && len(ve) > 0 {
		return formatValidationError(ve[0])
	}
	return err.Error()
}

func formatValidationError(fe validator.FieldError) string {
	name := jsonFieldLabel(fe.Field())
	switch fe.Tag() {
	case "required":
		return name + " is required"
	default:
		return fe.Error()
	}
}

// jsonFieldLabel 将结构体字段名转为 json 风格（TargetType → targetType）。
func jsonFieldLabel(field string) string {
	if idx := strings.LastIndex(field, "."); idx >= 0 {
		field = field[idx+1:]
	}
	if field == "" {
		return field
	}
	return strings.ToLower(field[:1]) + field[1:]
}

// currentAuthID 从 UserAuth/AdminAuth 注入的 auth_id 读取当前主体 ID；未登录返回 0。
func currentAuthID(c *gin.Context) uint {
	id, ok := c.Get("auth_id")
	if !ok {
		return 0
	}
	v, ok := id.(uint)
	if !ok {
		return 0
	}
	return v
}
