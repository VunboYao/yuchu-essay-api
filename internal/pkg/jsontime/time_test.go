package jsontime

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTime_MarshalJSON(t *testing.T) {
	tm := Time{Time: time.Date(2026, 6, 6, 23, 31, 32, 24000000, CST)}
	b, err := json.Marshal(tm)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := `"2026-06-06 23:31:32"`
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestTime_MarshalJSON_zero(t *testing.T) {
	b, err := json.Marshal(Time{})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "null" {
		t.Fatalf("got %s, want null", string(b))
	}
}

func TestTime_UnmarshalJSON_layout(t *testing.T) {
	var tm Time
	if err := json.Unmarshal([]byte(`"2026-06-06 23:31:32"`), &tm); err != nil {
		t.Fatal(err)
	}
	if tm.In(CST).Format(Layout) != "2026-06-06 23:31:32" {
		t.Fatalf("unexpected time: %v", tm)
	}
}
