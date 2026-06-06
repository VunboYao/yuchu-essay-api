// Package jsontime 提供 API JSON 与 GORM 共用的时间类型，统一输出东八区 yyyy-MM-dd HH:mm:ss。
package jsontime

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Layout 为 API 与可读字符串的统一格式（东八区）。
const Layout = "2006-01-02 15:04:05"

// CST 东八区，与 MySQL session 时区 +08:00 对齐。
var CST = time.FixedZone("CST", 8*3600)

// Time 包装 time.Time：数据库存 datetime，JSON 序列化为 Layout 字符串。
type Time struct {
	time.Time
}

// Now 返回当前时间的 jsontime.Time。
func Now() Time {
	return Time{Time: time.Now()}
}

// From 将标准 time.Time 转为 jsontime.Time。
func From(t time.Time) Time {
	return Time{Time: t}
}

// Ptr 返回指向 jsontime.Time 的指针，便于赋值可空字段。
func Ptr(t time.Time) *Time {
	tt := From(t)
	return &tt
}

// GormDataType 告知 GORM 使用 datetime 列类型。
func (Time) GormDataType() string {
	return "datetime"
}

// MarshalJSON 序列化为东八区 Layout 字符串；零值为 null。
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.In(CST).Format(Layout))
}

// UnmarshalJSON 支持 Layout 与 RFC3339 两种输入格式。
func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := parseTimeString(s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// Scan 从数据库读取 datetime 值。
func (t *Time) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	default:
		return fmt.Errorf("jsontime: cannot scan %T into Time", value)
	}
}

// Value 写入数据库 datetime 值。
func (t Time) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

func parseTimeString(s string) (time.Time, error) {
	if parsed, err := time.ParseInLocation(Layout, s, CST); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339, s); err == nil {
		return parsed, nil
	}
	return time.Time{}, fmt.Errorf("jsontime: cannot parse %q", s)
}
