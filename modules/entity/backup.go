package entity

import (
	errors2 "errors"
	"github.com/robfig/cron/v3"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type BackupState int

const (
	OnBackup      BackupState = iota + 1 // 备份中
	BackupFail                           // 备份失败
	BackupSuccess                        // 备份成功
)

func (bs BackupState) IsSuccess() bool {
	return bs == BackupSuccess
}

func (bs BackupState) IsOnBackup() bool {
	return bs == OnBackup
}

type BackupInfo struct {
	ID         int
	Note       string      `json:"note"` // 备注
	Name       string      `json:"name"` // 备份文件名
	BackupPath string      // 基于runtimePath的相对路径
	ShowPath   string      `json:"show_path"` // 客户端展示路径
	State      BackupState `json:"state"`
	CreatedAt  time.Time   `json:"created_at"` // 创建备份任务时间
	StartTime  time.Time   `json:"start_time"` // 执行备份任务时间
	Extensions string      // 备份数据的类型
	EntryID    cron.EntryID
}

func (bi *BackupInfo) TableName() string {
	return "backup_info"
}

func (bi *BackupInfo) OnBackup() bool {
	return bi.State == OnBackup
}

func CreateBackupInfo(info *BackupInfo) (err error) {
	err = GetDB().Create(info).Error
	if err != nil {
		err = errors.New(errors.InternalServerErr)
	}
	return
}

func UpdateBackupInfo(infoID int, updates map[string]interface{}) (err error) {
	if err = GetDB().First(&BackupInfo{}, "id = ?", infoID).Updates(updates).Error; err != nil {
		return
	}
	return
}

func DelBackupInfo(id int) (err error) {
	err = GetDB().Unscoped().Delete(&BackupInfo{ID: id}, id).Error
	return
}

func GetBackupInfo(id int) (info BackupInfo, err error) {
	err = GetDB().First(&info, "id = ?", id).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.BackupInfoNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

// GetLatestBackupInfo 获取最新的备份数据
func GetLatestBackupInfo() (*BackupInfo, error) {
	var backupInfo BackupInfo
	if err := GetDB().Order("backup_info.created_at DESC").First(&backupInfo).Error; err != nil {
		return nil, err
	}
	return &backupInfo, nil
}

func GetBackupList(page, size int) (backups []BackupInfo, err error) {
	db := GetDB()
	if page >= 0 && size > 0 {
		db = db.Limit(size).Offset(page)
	}
	if err = db.Order("created_at desc").Find(&backups).Error; err != nil {
		return
	}
	return
}
