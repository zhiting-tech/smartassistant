package supervisor

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/clouddisk"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/filebrowser"
	"os"
	"path/filepath"
)

const (
	restorePath = iota + 1
	backupID
)

type restoreReq struct {
	RestorePath     string   `json:"restore_path"`
	BackupID        int      `json:"backup_id"`
	RestoreType     int      `json:"restore_type"`
	PathType        PathType `json:"path_type"`
	RestoreFileName string   `json:"restore_file_name"`
	newRestorePath  string
}

func (req *restoreReq) validate(c *gin.Context) (err error) {
	switch req.RestoreType {
	case restorePath:
		if err = req.handlerRestorePath(c); err != nil{
			return
		}
	case backupID:
		// 对于不存在的文件，返回不存在提示，并且删除该数据记录
		var info entity.BackupInfo
		info, err = entity.GetBackupInfo(req.BackupID)
		if err != nil {
			return
		}
		if !info.State.IsSuccess() {
			err = errors.New(status.BackupInfoIsNotSuccessErr)
			return
		}
		_, err = os.Stat(filepath.Join(config.GetConf().SmartAssistant.RuntimePath, info.BackupPath, info.Name))
		if err != nil {
			if !os.IsNotExist(err) {
				err = errors.Wrap(err, errors.InternalServerErr)
				return
			}
			if err = entity.DelBackupInfo(info.ID); err != nil {
				return
			}
			err = errors.New(status.FileNotExistErr)
			return
		}
		req.newRestorePath = filepath.Join(info.BackupPath, info.Name)
	default:
		err = errors.Wrap(fmt.Errorf("validate err restore type %d", req.RestoreType), errors.BadRequest)
	}
	return
}

// Restore 启动恢复
func Restore(c *gin.Context) {
	var (
		req restoreReq
		err error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()
	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = req.validate(c); err != nil {
		return
	}
	err = supervisor.GetManager().StartRestoreJobWithContext(c.Request.Context(), req.newRestorePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Wrap(err, status.FileNotExistErr)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
}

func (req *restoreReq) handlerRestorePath(c *gin.Context) (err error) {
	switch req.PathType {
	case BackupPath:
		req.newRestorePath = req.RestorePath
	case Resource:
		accessToken := c.GetHeader(types.SATokenKey)
		var result clouddisk.GetResourcePathResp
		result, err = clouddisk.GetCloudDiskResourcePath(accessToken, req.RestorePath, c.Request.Context())
		if err != nil {
			return
		}
		req.newRestorePath = filepath.Join("volume", result.Data.Path)
	case MountedRes:
		req.newRestorePath = filepath.Join("volume", req.RestorePath)
	default:
		err = errors.Wrap(fmt.Errorf("handlerRestorePath err path_type is %d", req.PathType), errors.BadRequest)
		return
	}
	req.newRestorePath = filepath.Clean(req.newRestorePath)
	fs := filebrowser.GetFBOrInit()
	_, err = fs.Stat(req.newRestorePath)
	if err != nil {
		return
	}
	return
}
