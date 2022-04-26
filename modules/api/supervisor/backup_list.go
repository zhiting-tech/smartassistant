package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
)

const backupSizeDefault = 10

type ListBackupReq struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

// backupListResp 备份列表返回
type backupListResp struct {
	Backups []Backup `json:"backups"`
}

type Backup struct {
	BackupID   int    `json:"backup_id"`
	FileName   string `json:"file_name"`
	Note       string `json:"note"`
	CreatedAt  int64  `json:"created_at"`
	BackupPath string `json:"backup_path"`
	State      int    `json:"state"`
}

// ListBackup 备份列表
func ListBackup(c *gin.Context) {
	var (
		req  ListBackupReq
		resp backupListResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindQuery(&req); err != nil {
		return
	}
	if req.Size == 0 {
		req.Size = backupSizeDefault
	}
	backupList, err := entity.GetBackupList(req.Page, req.Size)
	if err != nil {
		return
	}

	resp.Backups = wrapResponse(backupList)
}

func wrapResponse(backups []entity.BackupInfo) []Backup {
	baks := make([]Backup, 0, len(backups))
	for _, b := range backups {
		baks = append(baks, Backup{
			BackupID:   b.ID,
			FileName:   b.Name,
			Note:       b.Note,
			CreatedAt:  b.StartTime.Unix(),
			BackupPath: b.ShowPath,
			State:      int(b.State),
		})
	}
	return baks
}
