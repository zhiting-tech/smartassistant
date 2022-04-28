package supervisor

import (
	errors2 "errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/zhiting-tech/smartassistant/modules/api/extension"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/clouddisk"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/backup"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type BackupAddReq struct {
	Note          string   `json:"note"`
	BackupPath    string   `json:"backup_path"`
	StartTime     int64    `json:"start_time"`
	Extensions    []string `json:"extensions"` // 扩展列表中的扩展名称
	PathType      PathType `json:"path_type"`
	name          string
	newBackupPath string
	showPath      string
}

func (req *BackupAddReq) validate(c *gin.Context) (err error) {
	// 目前只能同时只有一个备份进行
	latestInfo, err2 := entity.GetLatestBackupInfo()
	if err2 != nil {
		if !errors2.Is(err2, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err2, errors.InternalServerErr)
			return
		}
	}else {
		if latestInfo.State.IsOnBackup() {
			err = errors.New(status.OnBackupExist)
			return
		}
	}
	if req.StartTime > 0 {
		startTime := time.Unix(req.StartTime, 0)
		if startTime.Before(time.Now()) {
			err = errors.New(status.BackupTimeLessThenNowErr)
			return
		}
	}
	req.BackupPath = filepath.Clean(req.BackupPath)
	if req.BackupPath == "" || req.BackupPath == "/" {
		err = errors.Wrap(fmt.Errorf("path is %s", req.BackupPath), errors.BadRequest)
		return
	}
	if err = req.genNewBackupPath(c); err != nil {
		return
	}
	_, err = os.Stat(filepath.Join(config.GetConf().SmartAssistant.RuntimePath, req.newBackupPath))
	if os.IsNotExist(err) {
		err = errors.New(status.FileNotExistErr)
		return
	}
	if err = req.checkExtension(c); err != nil {
		return
	}
	return
}

func (req *BackupAddReq) checkExtension(c *gin.Context) (err error) {
	exNames := extension.GetExtensionsWithContext(c.Request.Context())
	for _, dt := range req.Extensions {
		hasExt := false
		for _, en := range exNames.ExtensionNames {
			if dt == en {
				hasExt = true
				break
			}
		}
		if !hasExt {
			err = errors.Wrap(fmt.Errorf("extension %s is not exist", req.Extensions), errors.BadRequest)
			return
		}
	}
	return
}

// genNewBackupPath 生成真实备份路径
func (req *BackupAddReq) genNewBackupPath(c *gin.Context) (err error) {
	switch req.PathType {
	case BackupPath:
		req.newBackupPath = req.BackupPath
		req.showPath = filepath.Join(InternalName, req.BackupPath)
	case Resource:
		accessToken := c.GetHeader(types.SATokenKey)
		var result clouddisk.GetResourcePathResp
		result, err = clouddisk.GetCloudDiskResourcePath(accessToken, req.BackupPath, c.Request.Context())
		if err != nil {
			return
		}
		req.newBackupPath = filepath.Join("volume", result.Data.Path)
		req.showPath = filepath.Join(InternalName, result.Data.ShowPath)
		return
	case MountedRes:
		req.newBackupPath = filepath.Join("volume", req.BackupPath)
		req.showPath = filepath.Join(ExternalName, req.BackupPath)
		return
	default:
		err = errors.Wrap(fmt.Errorf("gen path err path is %s, path_type is %d", req.BackupPath, req.PathType), errors.BadRequest)
		return
	}
	return
}

// AddBackup 创建并且启动备份
func AddBackup(c *gin.Context) {
	var (
		req     BackupAddReq
		err     error
		entryID cron.EntryID
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
	if err = req.validate(c); err != nil {
		return
	}
	req.name = fmt.Sprintf("backup-%s.zip", time.Now().Format("2006-01-02_15-04-05"))
	info := entity.BackupInfo{
		Note:       req.Note,
		CreatedAt:  time.Now(),
		State:      entity.OnBackup,
		StartTime:  time.Now(),
		BackupPath: req.newBackupPath,
		Name:       req.name,
		Extensions: strings.Join(req.Extensions, ","),
		ShowPath:   req.showPath,
	}
	err = entity.CreateBackupInfo(&info)
	updates := make(map[string]interface{})
	if err != nil {
		return
	}
	if req.StartTime > 0 {
		startJobTime := time.Unix(req.StartTime, 0)
		updates["start_time"] = startJobTime
		entryID, err = backup.AddBackupJob(startJobTime, info)
		if err != nil {
			logger.Warnf("add backup job error %v", err)
			updates["state"] = entity.BackupFail
		}
		updates["entry_id"] = entryID
		err = entity.UpdateBackupInfo(info.ID, updates)
		return
	}
	err = supervisor.GetManager().StartBackupJobWithContext(c.Request.Context(), info)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
}
