package supervisor

import (
	"strconv"
	"strings"
	"time"
)

func NewBackupFromFileName(fn string) Backup {
	var ca time.Time
	note := strings.TrimRight(fn, ".zip")
	sps := strings.Split(fn, "-")
	if len(sps) >= 2 {
		if tm, err := strconv.ParseInt(sps[0], 10, 64); err == nil {
			ca = time.Unix(tm, 0)
			note = strings.TrimSuffix(strings.Join(sps[1:], "-"), ".zip")
		}
	}
	return Backup{
		FileName:  fn,
		Note:      note,
		CreatedAt: ca,
	}
}

// Backup 备份描述文件结构 backup.json
type Backup struct {
	FileName  string    `json:"file_name"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}
