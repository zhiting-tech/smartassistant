package backup

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/job"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"os"
	"path/filepath"
	"time"
)

// CheckBackupInfo 检查如果是备份完后启动sa的，备份是否成功
func CheckBackupInfo() {
	latestData, err := entity.GetLatestBackupInfo()
	if err != nil {
		logger.Warnf("CheckBackupState error is %s", err)
		return
	}
	if !latestData.OnBackup() {
		return
	}
	// 对于大于当前时间的备份任务，将添加定时任务
	if latestData.EntryID > 0 && latestData.StartTime.After(time.Now()){
		entryID, err := AddBackupJob(latestData.StartTime, *latestData)
		if err != nil {
			logger.Warnf("sa start backup job err %s", err)
			return
		}
		if err = entity.UpdateBackupInfo(latestData.ID, map[string]interface{}{
			"entry_id": entryID,
		}); err != nil {
			logger.Warnf("update backup info state err %s", err)
			return
		}
		return
	}

	// 根据备份文件是否存在判断备份是否成功
	_, err = os.Stat(filepath.Join(config.GetConf().SmartAssistant.RuntimePath, latestData.BackupPath, latestData.Name))
	if err != nil {
		logger.Warnf("file stat err is %s", err)
		if !os.IsNotExist(err) {
			return
		}
		err2 := entity.UpdateBackupInfo(latestData.ID, map[string]interface{}{
			"state": entity.BackupFail,
		})
		if err2 != nil {
			logger.Warnf("update backup info state err %s", err2)
		}
		return
	}
	if err = entity.UpdateBackupInfo(latestData.ID, map[string]interface{}{
		"state": entity.BackupSuccess,
	}); err != nil {
		logger.Warnf("update backup info state err %s", err)
		return
	}
}

func AddBackupJob(t time.Time, info entity.BackupInfo) (entryID cron.EntryID, err error) {
	crontab := fmt.Sprintf("%d %d %d %d *", t.Minute(), t.Hour(), t.Day(), t.Month())
	entryID, err = job.GetJobServer().Cron.AddFunc(crontab, func() {
		err = supervisor.GetManager().StartBackupJobWithContext(context.Background(), info)
		if err != nil {
			logger.Warnf("run backup job error %v", err)
			info, err = entity.GetBackupInfo(info.ID)
			if err != nil {
				logger.Warnf("run backup job get info error %v", err)
				return
			}
			err = entity.UpdateBackupInfo(info.ID, map[string]interface{}{
				"state": entity.BackupFail,
			})
			logger.Warnf("run backup job update info error %v", err)
			job.GetJobServer().Cron.Remove(info.EntryID)
		}
	})
	return
}