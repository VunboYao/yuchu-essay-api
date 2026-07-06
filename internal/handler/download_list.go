// 下载记录列表：读时组装资料及所属专题展示摘要。
package handler

import (
	"course_server/internal/database"
	"course_server/internal/model"
	"course_server/internal/pkg/jsontime"

	"gorm.io/gorm"
)

// DownloadListItem 下载记录列表项。
type DownloadListItem struct {
	ID          uint          `json:"id"`
	FileID      uint          `json:"fileId"`
	OwnerType   string        `json:"ownerType"`
	OwnerID     uint          `json:"ownerId"`
	AccessType  string        `json:"accessType"`
	SourceScene string        `json:"sourceScene"`
	CreatedAt   jsontime.Time `json:"createdAt"`
	Available   bool          `json:"available"`
	Target      any           `json:"target"`
}

// DownloadFileSummary 资料下载记录卡片字段（不含 fileUrl）。
type DownloadFileSummary struct {
	FileName   string `json:"fileName"`
	FileType   string `json:"fileType"`
	FileSize   int64  `json:"fileSize"`
	OwnerTitle string `json:"ownerTitle,omitempty"`
	CoverImg   string `json:"coverImg,omitempty"`
}

type OwnerArticleRef struct {
	ID       uint   `gorm:"column:id"`
	Title    string `gorm:"column:title"`
	CoverImg string `gorm:"column:cover_img"`
}

func loadDownloadListPage(userID uint, page, pageSize int, resolve MediaURLResolver) ([]DownloadListItem, int64, error) {
	db := database.DB.Model(&model.DownloadLog{}).Where("user_id = ?", userID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []model.DownloadLog
	if err := paginate(db.Order("created_at DESC"), page, pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	if len(logs) == 0 {
		return []DownloadListItem{}, total, nil
	}
	files, owners, err := loadDownloadTargets(database.DB, logs)
	if err != nil {
		return nil, 0, err
	}
	return BuildDownloadListItems(logs, files, owners, resolve), total, nil
}

func loadDownloadTargets(db *gorm.DB, logs []model.DownloadLog) (
	map[uint]model.DownloadFile,
	map[uint]OwnerArticleRef,
	error,
) {
	fileIDs := uniqueUintIDs(collectDownloadFileIDs(logs))
	files := map[uint]model.DownloadFile{}
	if len(fileIDs) > 0 {
		var rows []model.DownloadFile
		if err := db.Select("id", "file_name", "file_type", "file_size", "tags", "access_type", "status", "owner_type", "owner_id").
			Where("id IN ?", fileIDs).Find(&rows).Error; err != nil {
			return nil, nil, err
		}
		for _, row := range rows {
			files[row.ID] = row
		}
	}
	ownerIDs := uniqueUintIDs(collectDownloadOwnerArticleIDs(files))
	owners := ownerArticleRefsByIDs(db, ownerIDs)
	return files, owners, nil
}

// DownloadFileListItem 后台下载资料列表项：资料字段 + 归属文章标题。
type DownloadFileListItem struct {
	model.DownloadFile
	OwnerTitle string `json:"ownerTitle,omitempty"`
}

// ownerArticleRefsByIDs 按文章 ID 批量查标题与封面，供列表读时组装。
func ownerArticleRefsByIDs(db *gorm.DB, ids []uint) map[uint]OwnerArticleRef {
	owners := map[uint]OwnerArticleRef{}
	ids = uniqueUintIDs(ids)
	if len(ids) == 0 {
		return owners
	}
	var rows []OwnerArticleRef
	db.Model(&model.Article{}).Select("id", "title", "cover_img").
		Where("id IN ?", ids).Find(&rows)
	for _, row := range rows {
		owners[row.ID] = row
	}
	return owners
}

func collectOwnerArticleIDsFromFiles(files []model.DownloadFile) []uint {
	ids := make([]uint, 0, len(files))
	for _, f := range files {
		if f.OwnerType == model.TargetTypeArticle && f.OwnerID > 0 {
			ids = append(ids, f.OwnerID)
		}
	}
	return ids
}

// BuildDownloadFileListItems 将下载资料与批量查到的归属文章标题合并为列表项。
func BuildDownloadFileListItems(files []model.DownloadFile, owners map[uint]OwnerArticleRef) []DownloadFileListItem {
	items := make([]DownloadFileListItem, 0, len(files))
	for _, f := range files {
		item := DownloadFileListItem{DownloadFile: f}
		if f.OwnerType == model.TargetTypeArticle && f.OwnerID > 0 {
			if owner, ok := owners[f.OwnerID]; ok {
				item.OwnerTitle = owner.Title
			}
		}
		items = append(items, item)
	}
	return items
}

func collectDownloadFileIDs(logs []model.DownloadLog) []uint {
	ids := make([]uint, 0, len(logs))
	for _, log := range logs {
		ids = append(ids, log.FileID)
	}
	return ids
}

func collectDownloadOwnerArticleIDs(files map[uint]model.DownloadFile) []uint {
	var ids []uint
	for _, file := range files {
		if file.OwnerType == model.TargetTypeArticle && file.OwnerID > 0 {
			ids = append(ids, file.OwnerID)
		}
	}
	return ids
}

func uniqueUintIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return nil
	}
	seen := map[uint]bool{}
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// BuildDownloadListItems 将下载记录与批量查到的资料合并为列表项。
func BuildDownloadListItems(
	logs []model.DownloadLog,
	files map[uint]model.DownloadFile,
	owners map[uint]OwnerArticleRef,
	resolve MediaURLResolver,
) []DownloadListItem {
	items := make([]DownloadListItem, 0, len(logs))
	for _, log := range logs {
		item := DownloadListItem{
			ID:          log.ID,
			FileID:      log.FileID,
			OwnerType:   log.OwnerType,
			OwnerID:     log.OwnerID,
			AccessType:  log.AccessType,
			SourceScene: log.SourceScene,
			CreatedAt:   log.CreatedAt,
		}
		file, ok := files[log.FileID]
		if !ok {
			item.Available = false
			item.Target = nil
		} else {
			item.Available = model.IsPublicFreeDownload(file)
			summary := DownloadFileSummary{
				FileName: file.FileName,
				FileType: file.FileType,
				FileSize: file.FileSize,
			}
			if file.OwnerType == model.TargetTypeArticle && file.OwnerID > 0 {
				if owner, ok := owners[file.OwnerID]; ok {
					summary.OwnerTitle = owner.Title
					summary.CoverImg = owner.CoverImg
					if resolve != nil {
						summary.CoverImg = resolve(summary.CoverImg)
					}
				}
			}
			item.Target = summary
		}
		items = append(items, item)
	}
	return items
}
