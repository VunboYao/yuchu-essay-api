// 下载资料 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package model

import (
	"fmt"
	"strings"
)

// DownloadFileWriteRequest 下载资料创建/更新共用入参；不含 id、时间戳与 downloadCount。
type DownloadFileWriteRequest struct {
	OwnerType  string `json:"ownerType" binding:"required"`
	OwnerID    uint   `json:"ownerId" binding:"required"`
	FileName   string `json:"fileName" binding:"required"`
	Tags       string `json:"tags"`
	AccessType string `json:"accessType" binding:"required"`
	FileType   string `json:"fileType"`
	FileURL    string `json:"fileUrl" binding:"required"`
	FileSize   int64  `json:"fileSize"`
	NeedLogin  bool   `json:"needLogin"`
	SortOrder  int    `json:"sortOrder"`
	Status     *int   `json:"status"`
}

// CreateDownloadFileRequest 创建下载资料请求。
type CreateDownloadFileRequest = DownloadFileWriteRequest

// UpdateDownloadFileRequest 更新下载资料请求。
type UpdateDownloadFileRequest = DownloadFileWriteRequest

func validAccessType(accessType string) bool {
	switch accessType {
	case AccessTypeFree, AccessTypePurchaseRequired, AccessTypeMemberOnly, AccessTypeInternal:
		return true
	default:
		return false
	}
}

func (r DownloadFileWriteRequest) normalize() (DownloadFile, error) {
	ownerType := strings.TrimSpace(r.OwnerType)
	if ownerType == "" {
		return DownloadFile{}, fmt.Errorf("ownerType is required")
	}
	if r.OwnerID == 0 {
		return DownloadFile{}, fmt.Errorf("ownerId is required")
	}
	fileName := strings.TrimSpace(r.FileName)
	if fileName == "" {
		return DownloadFile{}, fmt.Errorf("fileName is required")
	}
	accessType := strings.TrimSpace(r.AccessType)
	if accessType == "" {
		return DownloadFile{}, fmt.Errorf("accessType is required")
	}
	if !validAccessType(accessType) {
		return DownloadFile{}, fmt.Errorf("invalid accessType")
	}
	fileURL := strings.TrimSpace(r.FileURL)
	if fileURL == "" {
		return DownloadFile{}, fmt.Errorf("fileUrl is required")
	}
	return DownloadFile{
		OwnerType:  ownerType,
		OwnerID:    r.OwnerID,
		FileName:   fileName,
		Tags:       strings.TrimSpace(r.Tags),
		AccessType: accessType,
		FileType:   strings.TrimSpace(r.FileType),
		FileURL:    fileURL,
		FileSize:   r.FileSize,
		NeedLogin:  r.NeedLogin,
		SortOrder:  r.SortOrder,
	}, nil
}

// ValidateForCreate 校验创建入参；status 未传时默认启用。
func (r DownloadFileWriteRequest) ValidateForCreate() (DownloadFile, error) {
	file, err := r.normalize()
	if err != nil {
		return DownloadFile{}, err
	}
	status, err := ResolveStatus(r.Status, StatusEnabled)
	if err != nil {
		return DownloadFile{}, err
	}
	file.Status = status
	return file, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r DownloadFileWriteRequest) ValidateForUpdate(existing DownloadFile) (DownloadFile, error) {
	file, err := r.normalize()
	if err != nil {
		return DownloadFile{}, err
	}
	status, err := ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return DownloadFile{}, err
	}
	file.Status = status
	return file, nil
}

// UpdateMap 返回 GORM Updates 用的字段 map（支持零值写入；不含 downloadCount）。
func (f DownloadFile) UpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"owner_type":  f.OwnerType,
		"owner_id":    f.OwnerID,
		"file_name":   f.FileName,
		"tags":        f.Tags,
		"access_type": f.AccessType,
		"file_type":   f.FileType,
		"file_url":    f.FileURL,
		"file_size":   f.FileSize,
		"need_login":  f.NeedLogin,
		"sort_order":  f.SortOrder,
		"status":      f.Status,
	}
}
