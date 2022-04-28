package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/job"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"os"
	"path/filepath"
)

type backupDelReq struct {
	BackupID int `json:"backup_id"`
}

// DeleteBackup 删除备份
func DeleteBackup(c *gin.Context) {
	var (
		req backupDelReq
		err error
		fi  os.FileInfo
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()
	err = c.BindJSON(&req)
	if err != nil {
		logger.Warnf("request error %v", err)
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	info, err := entity.GetBackupInfo(req.BackupID)
	if err != nil {
		return
	}
	backupFilePath := filepath.Join(config.GetConf().SmartAssistant.RuntimePath, info.BackupPath, info.Name)
	if info.State.IsSuccess() {
		fi, err = os.Stat(backupFilePath)
		if err != nil {
			if !os.IsNotExist(err) {
				err = errors.Wrap(err, errors.InternalServerErr)
				return
			}
		}
		if err == nil && !fi.IsDir() {
			if err = os.RemoveAll(backupFilePath); err != nil {
				return
			}
		}
	}
	if info.State.IsOnBackup() {
		job.GetJobServer().Cron.Remove(info.EntryID)
	}
	if err = entity.DelBackupInfo(info.ID); err != nil {
		return
	}
}
