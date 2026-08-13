// Package download 小程序免费下载签名与激励解锁用例（handler 只做 HTTP）。
package download

import (
	"context"
	"errors"
	"strings"

	"course_server/internal/model"
	"course_server/internal/pkg/adunlock"

	"gorm.io/gorm"
)

const signURLExpiresIn = 300

// URLSigner 签发下载地址；生产注入 *ossutil.Client，测试注入假实现。
type URLSigner interface {
	SignDownloadURL(fileURL string, expireSec int64) string
}

// Service 下载签名 / 激励解锁；禁止 import gin。
type Service struct {
	db        *gorm.DB
	jwtSecret string
	jwtIssuer string
	signer    URLSigner
}

func New(db *gorm.DB, jwtSecret, jwtIssuer string, signer URLSigner) *Service {
	return &Service{db: db, jwtSecret: jwtSecret, jwtIssuer: jwtIssuer, signer: signer}
}

type SignInput struct {
	FileID      uint
	UserID      uint
	UnlockToken string
	SourceScene string
}

type SignResult struct {
	URL       string
	ExpiresIn int
	File      model.DownloadFile
}

type UnlockInput struct {
	FileID   uint
	UserID   uint
	PagePath string
}

type UnlockResult struct {
	UnlockToken string
	ExpiresIn   int
}

// SignDownloadURL 校验免费三条件、needLogin、激励门槛后记日志/计数并签发 URL。
func (s *Service) SignDownloadURL(ctx context.Context, in SignInput) (*SignResult, error) {
	file, err := s.loadFile(ctx, in.FileID)
	if err != nil {
		return nil, err
	}
	if err := checkFreeAndLogin(file, in.UserID); err != nil {
		return nil, err
	}
	gate, err := s.loadDownloadGate(ctx)
	if err != nil {
		return nil, err
	}
	// 门槛开启时 token 须同时匹配 fileId、当前 userId、当前 gate.slotKey
	if gate.IsDownloadGateActive() {
		token := strings.TrimSpace(in.UnlockToken)
		if token == "" {
			return nil, ErrRewardedVideoRequired
		}
		claims, err := adunlock.Verify(s.jwtSecret, token, file.ID)
		if err != nil || claims.UserID != in.UserID || claims.SlotKey != gate.SlotKey {
			return nil, ErrInvalidUnlockToken
		}
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.DownloadLog{
			UserID:      in.UserID,
			FileID:      file.ID,
			OwnerType:   file.OwnerType,
			OwnerID:     file.OwnerID,
			AccessType:  file.AccessType,
			SourceScene: in.SourceScene,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&file).UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error; err != nil {
			return err
		}
		if file.OwnerType == model.TargetTypeArticle {
			if err := tx.Model(&model.Article{}).Where("id = ?", file.OwnerID).UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, ErrRecordDownload
	}

	return &SignResult{
		URL:       s.signer.SignDownloadURL(file.FileURL, signURLExpiresIn),
		ExpiresIn: signURLExpiresIn,
		File:      file,
	}, nil
}

// UnlockDownloadAd 门槛启用时记录激励完成并签发短时 unlockToken。
func (s *Service) UnlockDownloadAd(ctx context.Context, in UnlockInput) (*UnlockResult, error) {
	file, err := s.loadFile(ctx, in.FileID)
	if err != nil {
		return nil, err
	}
	if err := checkFreeAndLogin(file, in.UserID); err != nil {
		return nil, err
	}
	gate, err := s.loadDownloadGate(ctx)
	if err != nil {
		return nil, err
	}
	if !gate.IsDownloadGateActive() {
		return nil, ErrGateNotEnabled
	}

	if err := s.db.WithContext(ctx).Create(&model.AdEvent{
		UserID:     in.UserID,
		SlotKey:    gate.SlotKey,
		EventType:  model.AdEventRewardedComplete,
		PagePath:   in.PagePath,
		TargetType: model.TargetTypeDownloadFile,
		TargetID:   file.ID,
	}).Error; err != nil {
		return nil, ErrRecordAdUnlock
	}

	token, expiresIn, err := adunlock.Sign(s.jwtSecret, s.jwtIssuer, file.ID, in.UserID, gate.SlotKey)
	if err != nil {
		return nil, ErrIssueToken
	}
	return &UnlockResult{UnlockToken: token, ExpiresIn: expiresIn}, nil
}

func (s *Service) loadFile(ctx context.Context, fileID uint) (model.DownloadFile, error) {
	var file model.DownloadFile
	if err := s.db.WithContext(ctx).First(&file, fileID).Error; err != nil {
		return model.DownloadFile{}, ErrNotFound
	}
	return file, nil
}

// checkFreeAndLogin 一期底线：免费三条件通过后才看 needLogin（未登录 userID=0）。
func checkFreeAndLogin(file model.DownloadFile, userID uint) error {
	if !model.IsPublicFreeDownload(file) {
		return ErrNotFree
	}
	if file.NeedLogin && userID == 0 {
		return ErrLoginRequired
	}
	return nil
}

func (s *Service) loadDownloadGate(ctx context.Context) (model.AdSlot, error) {
	slot, err := model.FindDownloadGateSlot(s.db.WithContext(ctx))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.AdSlot{}, nil
	}
	if err != nil {
		return model.AdSlot{}, ErrGateCheck
	}
	return slot, nil
}
