// 下载资料 Admin 写操作 Request DTO 与校验（见 AGENTS.md §6）。
package admin

import (
	"fmt"
	"strings"

	"course_server/internal/dto"
	"course_server/internal/model"
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

// validOwnerType 一期仅允许挂在作文专题文章下；二期扩展 owner 类型时在此追加。
func validOwnerType(ownerType string) bool {
	switch ownerType {
	case model.TargetTypeArticle:
		return true
	default:
		return false
	}
}

func validAccessType(accessType string) bool {
	switch accessType {
	case model.AccessTypeFree, model.AccessTypePurchaseRequired, model.AccessTypeMemberOnly, model.AccessTypeInternal:
		return true
	default:
		return false
	}
}

func (r DownloadFileWriteRequest) normalize() (model.DownloadFile, error) {
	ownerType := strings.TrimSpace(r.OwnerType)
	if ownerType == "" {
		return model.DownloadFile{}, fmt.Errorf("ownerType is required")
	}
	if r.OwnerID == 0 {
		return model.DownloadFile{}, fmt.Errorf("ownerId is required")
	}
	if !validOwnerType(ownerType) {
		return model.DownloadFile{}, fmt.Errorf("invalid ownerType")
	}
	fileName := strings.TrimSpace(r.FileName)
	if fileName == "" {
		return model.DownloadFile{}, fmt.Errorf("fileName is required")
	}
	accessType := strings.TrimSpace(r.AccessType)
	if accessType == "" {
		return model.DownloadFile{}, fmt.Errorf("accessType is required")
	}
	if !validAccessType(accessType) {
		return model.DownloadFile{}, fmt.Errorf("invalid accessType")
	}
	fileURL := strings.TrimSpace(r.FileURL)
	if fileURL == "" {
		return model.DownloadFile{}, fmt.Errorf("fileUrl is required")
	}
	return model.DownloadFile{
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
func (r DownloadFileWriteRequest) ValidateForCreate() (model.DownloadFile, error) {
	file, err := r.normalize()
	if err != nil {
		return model.DownloadFile{}, err
	}
	status, err := dto.ResolveStatus(r.Status, model.StatusEnabled)
	if err != nil {
		return model.DownloadFile{}, err
	}
	file.Status = status
	return file, nil
}

// ValidateForUpdate 校验更新入参；status 未传时保留 existing 原值。
func (r DownloadFileWriteRequest) ValidateForUpdate(existing model.DownloadFile) (model.DownloadFile, error) {
	file, err := r.normalize()
	if err != nil {
		return model.DownloadFile{}, err
	}
	status, err := dto.ResolveStatus(r.Status, existing.Status)
	if err != nil {
		return model.DownloadFile{}, err
	}
	file.Status = status
	return file, nil
}
