package setting

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type GetSettingResp struct {
	UserCredentialFoundSetting    entity.UserCredentialFoundSetting    `json:"user_credential_found_setting"`
	LogSetting                    entity.LogSwitchSetting              `json:"log_setting"`
	GetCloudDiskCredentialSetting entity.GetCloudDiskCredentialSetting `json:"get_cloud_disk_credential_setting"`
}

// GetSetting 获取全局配置
func GetSetting(c *gin.Context) {

	var (
		resp GetSettingResp
		err  error
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	globalSettings, err := entity.GetAllSetting(session.Get(c).AreaID)
	if err != nil {
		return
	}

	for _, gs := range globalSettings {
		switch gs.Type {
		case entity.UserCredentialFoundType:
			err = json.Unmarshal(gs.Value, &resp.UserCredentialFoundSetting)
		case entity.LogSwitch:
			err = json.Unmarshal(gs.Value, &resp.LogSetting)
		case entity.GetCloudDiskCredential:
			err = json.Unmarshal(gs.Value, &resp.GetCloudDiskCredentialSetting)
		}
		if err != nil {
			logger.Errorf("json unmarshal err", err)
			err = errors.New(errors.InternalServerErr)
			return
		}
	}

}
