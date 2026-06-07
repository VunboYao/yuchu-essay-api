// HTTP 入参通用校验（见 AGENTS.md §6）。
package dto

import (
	"fmt"

	"course_server/internal/model"
)

// ResolveStatus 解析可选 status：nil 时用 defaultStatus；非 nil 时须为 StatusDisabled 或 StatusEnabled。
func ResolveStatus(ptr *int, defaultStatus int) (int, error) {
	if ptr == nil {
		return defaultStatus, nil
	}
	if *ptr != model.StatusDisabled && *ptr != model.StatusEnabled {
		return 0, fmt.Errorf("invalid status")
	}
	return *ptr, nil
}
