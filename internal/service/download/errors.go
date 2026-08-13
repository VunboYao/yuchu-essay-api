package download

import "errors"

// 对外稳定文案：handler 按 sentinel 映射 HTTP，禁止把底层 err.Error() 拼进 500。
var (
	ErrNotFound              = errors.New("download file not found")
	ErrNotFree               = errors.New("download file is not free")
	ErrLoginRequired         = errors.New("login required")
	ErrRewardedVideoRequired = errors.New("rewarded video required")
	ErrInvalidUnlockToken    = errors.New("invalid or expired unlock token")
	ErrGateNotEnabled        = errors.New("download rewarded video gate is not enabled")
	ErrGateCheck             = errors.New("failed to check download gate")
	ErrRecordDownload        = errors.New("failed to record download")
	ErrRecordAdUnlock        = errors.New("failed to record ad unlock")
	ErrIssueToken            = errors.New("failed to issue unlock token")
)
