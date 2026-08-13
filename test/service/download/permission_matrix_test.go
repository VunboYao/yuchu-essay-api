package download_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"course_server/internal/model"
	"course_server/internal/pkg/adunlock"
	"course_server/internal/service/download"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	testJWTSecret = "test-secret"
	testJWTIssuer = "essay-mvp-test"
)

type stubSigner struct{}

func (stubSigner) SignDownloadURL(fileURL string, expireSec int64) string {
	return fileURL + "?signed=1&exp=" + strconv.FormatInt(expireSec, 10)
}

func setupService(t *testing.T) (*gorm.DB, *download.Service) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Article{},
		&model.DownloadFile{},
		&model.AdSlot{},
		&model.DownloadLog{},
		&model.AdEvent{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db, download.New(db, testJWTSecret, testJWTIssuer, stubSigner{})
}

func seedFile(t *testing.T, db *gorm.DB, overrides func(*model.DownloadFile)) model.DownloadFile {
	t.Helper()
	file := model.DownloadFile{
		OwnerType:  model.TargetTypeArticle,
		OwnerID:    1,
		FileName:   "测试资料.pdf",
		Tags:       "免费",
		AccessType: model.AccessTypeFree,
		FileType:   "pdf",
		FileURL:    "https://oss.example.com/test.pdf",
		Status:     model.StatusEnabled,
		NeedLogin:  false,
	}
	if overrides != nil {
		overrides(&file)
	}
	needLogin := file.NeedLogin
	status := file.Status
	if err := db.Create(&file).Error; err != nil {
		t.Fatalf("seed file: %v", err)
	}
	if err := db.Model(&file).Updates(map[string]interface{}{
		"need_login": needLogin,
		"status":     status,
	}).Error; err != nil {
		t.Fatalf("seed file flags: %v", err)
	}
	file.NeedLogin = needLogin
	file.Status = status
	return file
}

func seedGate(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Create(&model.AdSlot{
		SlotKey:  model.AdSlotKeyDownloadRewardedVideo,
		SlotName: "下载激励视频",
		Page:     model.AdPageDownload,
		AdType:   model.AdTypeRewardedVideo,
		AdUnitID: "adunit-test",
		Status:   model.StatusEnabled,
	}).Error; err != nil {
		t.Fatalf("seed gate: %v", err)
	}
}

func signToken(t *testing.T, fileID, userID uint, slotKey string) string {
	t.Helper()
	token, _, err := adunlock.Sign(testJWTSecret, testJWTIssuer, fileID, userID, slotKey)
	if err != nil {
		t.Fatalf("sign unlock: %v", err)
	}
	return token
}

func TestSignDownloadURLPermissionMatrix(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name    string
		setup   func(t *testing.T, db *gorm.DB) download.SignInput
		wantErr error
	}{
		{
			name: "not found",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				return download.SignInput{FileID: 999}
			},
			wantErr: download.ErrNotFound,
		},
		{
			name: "disabled not free",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, func(f *model.DownloadFile) {
					f.Status = model.StatusDisabled
				})
				return download.SignInput{FileID: file.ID}
			},
			wantErr: download.ErrNotFree,
		},
		{
			name: "missing free tag",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, func(f *model.DownloadFile) {
					f.Tags = "PDF"
				})
				return download.SignInput{FileID: file.ID}
			},
			wantErr: download.ErrNotFree,
		},
		{
			name: "purchase_required",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, func(f *model.DownloadFile) {
					f.AccessType = model.AccessTypePurchaseRequired
				})
				return download.SignInput{FileID: file.ID}
			},
			wantErr: download.ErrNotFree,
		},
		{
			name: "needLogin anonymous",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, func(f *model.DownloadFile) {
					f.NeedLogin = true
				})
				return download.SignInput{FileID: file.ID, UserID: 0}
			},
			wantErr: download.ErrLoginRequired,
		},
		{
			name: "needLogin with user",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, func(f *model.DownloadFile) {
					f.NeedLogin = true
				})
				return download.SignInput{FileID: file.ID, UserID: 7}
			},
		},
		{
			name: "anonymous without gate",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, nil)
				return download.SignInput{FileID: file.ID}
			},
		},
		{
			name: "gate without token",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, nil)
				seedGate(t, db)
				return download.SignInput{FileID: file.ID}
			},
			wantErr: download.ErrRewardedVideoRequired,
		},
		{
			name: "gate junk token",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, nil)
				seedGate(t, db)
				return download.SignInput{FileID: file.ID, UnlockToken: "not-a-jwt"}
			},
			wantErr: download.ErrInvalidUnlockToken,
		},
		{
			name: "gate userId mismatch",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, nil)
				seedGate(t, db)
				token := signToken(t, file.ID, 8, model.AdSlotKeyDownloadRewardedVideo)
				return download.SignInput{FileID: file.ID, UserID: 7, UnlockToken: token}
			},
			wantErr: download.ErrInvalidUnlockToken,
		},
		{
			name: "gate slotKey mismatch",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, nil)
				seedGate(t, db)
				token := signToken(t, file.ID, 0, "other_slot")
				return download.SignInput{FileID: file.ID, UnlockToken: token}
			},
			wantErr: download.ErrInvalidUnlockToken,
		},
		{
			name: "gate fileId mismatch",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, nil)
				seedGate(t, db)
				token := signToken(t, file.ID+1, 0, model.AdSlotKeyDownloadRewardedVideo)
				return download.SignInput{FileID: file.ID, UnlockToken: token}
			},
			wantErr: download.ErrInvalidUnlockToken,
		},
		{
			name: "gate matched token",
			setup: func(t *testing.T, db *gorm.DB) download.SignInput {
				file := seedFile(t, db, nil)
				seedGate(t, db)
				token := signToken(t, file.ID, 7, model.AdSlotKeyDownloadRewardedVideo)
				return download.SignInput{FileID: file.ID, UserID: 7, UnlockToken: token}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, svc := setupService(t)
			in := tc.setup(t, db)
			out, err := svc.SignDownloadURL(ctx, in)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err=%v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if out == nil || out.URL == "" || out.ExpiresIn != 300 {
				t.Fatalf("unexpected result: %+v", out)
			}
		})
	}
}

func TestUnlockDownloadAdNeedLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("anonymous", func(t *testing.T) {
		db, svc := setupService(t)
		file := seedFile(t, db, func(f *model.DownloadFile) {
			f.NeedLogin = true
		})
		seedGate(t, db)
		_, err := svc.UnlockDownloadAd(ctx, download.UnlockInput{FileID: file.ID})
		if !errors.Is(err, download.ErrLoginRequired) {
			t.Fatalf("err=%v, want %v", err, download.ErrLoginRequired)
		}
	})

	t.Run("logged in", func(t *testing.T) {
		db, svc := setupService(t)
		file := seedFile(t, db, func(f *model.DownloadFile) {
			f.NeedLogin = true
		})
		seedGate(t, db)
		out, err := svc.UnlockDownloadAd(ctx, download.UnlockInput{FileID: file.ID, UserID: 7})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if out.UnlockToken == "" || out.ExpiresIn != 300 {
			t.Fatalf("unexpected result: %+v", out)
		}
	})
}
