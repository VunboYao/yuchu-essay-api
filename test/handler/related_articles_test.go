package handler_test

import (
	"testing"

	"course_server/internal/handler"
)

func TestRelatedArticleMatches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		baseGrade   int
		baseSemester string
		baseUnit    int
		grade       int
		semester    string
		unit        int
		want        bool
	}{
		{"同年级同册次同单元", 6, "上册", 2, 6, "上册", 2, true},
		{"年级不同", 6, "上册", 2, 5, "上册", 2, false},
		{"册次不同", 6, "上册", 2, 6, "下册", 2, false},
		{"单元不同", 6, "上册", 2, 6, "上册", 1, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := handler.RelatedArticleMatches(
				tc.baseGrade, tc.baseSemester, tc.baseUnit,
				tc.grade, tc.semester, tc.unit,
			)
			if got != tc.want {
				t.Fatalf("match = %v, want %v", got, tc.want)
			}
		})
	}
}
