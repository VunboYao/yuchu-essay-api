package jsontime

import (
	"encoding/json"
	"testing"
	"time"
)

// 验证 model 嵌入场景下 JSON 输出符合 API 约定。
func TestTime_modelJSON(t *testing.T) {
	type sample struct {
		CreatedAt Time `json:"createdAt"`
	}
	s := sample{CreatedAt: Time{Time: time.Date(2026, 6, 6, 23, 31, 32, 0, CST)}}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"createdAt":"2026-06-06 23:31:32"}`
	if string(b) != want {
		t.Fatalf("got %s, want %s", string(b), want)
	}
}
