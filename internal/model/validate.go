package model

import "fmt"

// ResolveStatus 解析可选 status：nil 时用 defaultStatus；非 nil 时须为 StatusDisabled 或 StatusEnabled。
func ResolveStatus(ptr *int, defaultStatus int) (int, error) {
	if ptr == nil {
		return defaultStatus, nil
	}
	if *ptr != StatusDisabled && *ptr != StatusEnabled {
		return 0, fmt.Errorf("invalid status")
	}
	return *ptr, nil
}
