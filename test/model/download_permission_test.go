package model_test

import (
	"testing"

	"course_server/internal/model"
)

func TestIsPublicFreeDownload(t *testing.T) {
	t.Parallel()

	freeFile := model.DownloadFile{
		Status:     model.StatusEnabled,
		AccessType: model.AccessTypeFree,
		Tags:       "免费,PDF",
	}

	cases := []struct {
		name string
		file model.DownloadFile
		want bool
	}{
		{name: "enabled free with tag", file: freeFile, want: true},
		{name: "disabled", file: func() model.DownloadFile {
			f := freeFile
			f.Status = model.StatusDisabled
			return f
		}(), want: false},
		{name: "missing free tag", file: func() model.DownloadFile {
			f := freeFile
			f.Tags = "PDF"
			return f
		}(), want: false},
		{name: "purchase_required", file: func() model.DownloadFile {
			f := freeFile
			f.AccessType = model.AccessTypePurchaseRequired
			return f
		}(), want: false},
		{name: "member_only", file: func() model.DownloadFile {
			f := freeFile
			f.AccessType = model.AccessTypeMemberOnly
			return f
		}(), want: false},
		{name: "internal", file: func() model.DownloadFile {
			f := freeFile
			f.AccessType = model.AccessTypeInternal
			return f
		}(), want: false},
		{name: "empty tags", file: func() model.DownloadFile {
			f := freeFile
			f.Tags = ""
			return f
		}(), want: false},
		{name: "tag substring match", file: func() model.DownloadFile {
			f := freeFile
			f.Tags = "资料,免费下载"
			return f
		}(), want: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := model.IsPublicFreeDownload(tc.file); got != tc.want {
				t.Fatalf("IsPublicFreeDownload() = %v, want %v (file=%+v)", got, tc.want, tc.file)
			}
		})
	}
}
