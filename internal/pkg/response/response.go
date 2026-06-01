// Package response 统一 HTTP JSON 响应格式：{ code, message, data }（Day1）。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// Success 业务成功，HTTP 200 且 code=0。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: "success", Data: data})
}

// Page 分页列表，data 内为 { list, total, page, pageSize }。
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func Fail(c *gin.Context, status int, code int, message string) {
	c.JSON(status, Body{Code: code, Message: message})
}

func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, 400, message)
}

func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, 401, message)
}

func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, 403, message)
}

func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, 404, message)
}

func ServerError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, 500, message)
}
