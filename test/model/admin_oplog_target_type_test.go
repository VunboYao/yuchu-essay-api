package model_test

import (
	"testing"

	"course_server/internal/model"
)

func TestAdminOpLogTargetTypeFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/admin/api/v1/articles", model.TargetTypeArticle},
		{"/admin/api/v1/articles/:id", model.TargetTypeArticle},
		{"/admin/api/v1/essay-examples/:id", model.TargetTypeEssay},
		{"/admin/api/v1/download-files", model.TargetTypeDownloadFile},
		{"/admin/api/v1/tags/:id", model.AdminTargetTypeTag},
		{"/admin/api/v1/taxonomy-units/:id", model.AdminTargetTypeTaxonomyUnit},
		{"/admin/api/v1/home-blocks/:id", model.AdminTargetTypeHomeBlock},
		{"/admin/api/v1/videos/:id", model.AdminTargetTypeVideo},
		{"/admin/api/v1/video-channel", model.AdminTargetTypeVideoChannel},
		{"/admin/api/v1/ad-slots", model.AdminTargetTypeAdSlot},
		{"/admin/api/v1/official-account-configs", model.AdminTargetTypeOfficialAccount},
		{"/admin/api/v1/upload/sts", model.AdminTargetTypeUpload},
		{"/api/v1/articles", ""},
	}
	for _, tt := range tests {
		if got := model.AdminOpLogTargetTypeFromPath(tt.path); got != tt.want {
			t.Errorf("AdminOpLogTargetTypeFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
