// 分页与鉴权上下文辅助函数，供各 handler 复用。
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

func paginate(db *gorm.DB, page, pageSize int) *gorm.DB {
	return db.Offset((page - 1) * pageSize).Limit(pageSize)
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
