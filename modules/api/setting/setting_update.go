package setting

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/logreplay"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type RemoteHelpSetting struct {
	Enable bool `json:"enable"`
}

type GetSettingQuery struct {
	UserCredentialFoundSetting    *entity.UserCredentialFoundSetting    `json:"user_credential_found_setting"`
	LogSetting                    *entity.LogSwitchSetting              `json:"log_setting"`
	GetCloudDiskCredentialSetting *entity.GetCloudDiskCredentialSetting `json:"get_cloud_disk_credential_setting"`
	RemoteHelp                    *RemoteHelpSetting                    `json:"remote_help"`
}

// UpdateSetting 修改全局配置
func UpdateSetting(c *gin.Context) {

	var (
		req         GetSettingQuery
		err         error
		sessionUser *session.User
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	// 只有SA拥有者才能设置
	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	// 修改是否允许找回用户凭证的配置
	if req.UserCredentialFoundSetting != nil {
		err = req.UpdateUserCredentialFound(sessionUser.AreaID)
		if err != nil {
			return
		}
	}

	// 修改是否打开日志上传
	if req.LogSetting != nil {
		err = req.UpdateLogSetting(sessionUser.AreaID)
		if err != nil {
			return
		}
	}

	// 修改是否允许获取网盘凭证的配置
	if req.GetCloudDiskCredentialSetting != nil {
		if err = req.UpdateCloudDiskCredential(sessionUser.AreaID); err != nil {
			return
		}
	}

	// 是否开启远程协助
	if req.RemoteHelp != nil {
		if err = req.UpdateRemoteHelpWithContext(c.Request.Context()); err != nil {
			return
		}
	}
}

func (req *GetSettingQuery) UpdateUserCredentialFound(areaID uint64) (err error) {

	// 更新是否允许找回用户凭证
	setting := req.UserCredentialFoundSetting
	err = entity.UpdateSetting(entity.UserCredentialFoundType, &setting, areaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	// 发送SC请求SA的认证token给SC
	if setting.UserCredentialFound {
		go sendAreaAuthToSC(areaID)
	}

	return nil
}

func (req *GetSettingQuery) UpdateLogSetting(areaID uint64) (err error) {
	setting := req.LogSetting
	err = entity.UpdateSetting(entity.LogSwitch, &setting, areaID)

	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if setting.LogSwitch {
		logreplay.GetLogPlayer().EnableUpload()
	} else {
		logreplay.GetLogPlayer().DisableUpload()
	}

	return nil
}

func (req *GetSettingQuery) UpdateCloudDiskCredential(areaID uint64) (err error) {
	// 更新是否允许获取网盘凭证
	setting := req.GetCloudDiskCredentialSetting
	err = entity.UpdateSetting(entity.GetCloudDiskCredential, &setting, areaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 发送sc请求sa的认证token给SC
	if setting.GetCloudDiskCredentialSetting {
		go sendAreaAuthToSC(areaID)
	}
	return
}

func (req *GetSettingQuery) UpdateRemoteHelpWithContext(ctx context.Context) (err error) {
	if req.RemoteHelp.Enable {
		var privateKey, publicKey []byte
		if privateKey, publicKey, err = sshKeyGen(); err != nil {
			return
		}
		if err = SendPrivateKeyToSCWithContext(ctx, privateKey); err != nil {
			return
		}
		if err = supervisor.GetClient().EnableRemoteHelpWithContext(ctx, publicKey); err != nil {
			return
		}
	} else {
		err = supervisor.GetClient().DisableRemoteHelpWithContext(ctx)
	}

	return
}
