package device

import (
	errors2 "errors"
	"os"

	"github.com/zhiting-tech/smartassistant/modules/types"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

// CheckSaDeviceResp 检查SA设备绑定情况接口请求参数
type CheckSaDeviceResp struct {
	IsBind     bool   `json:"is_bind"`
	Revision   string `json:"revision"`
	Version    string `json:"version"`
	MinVersion string `json:"min_version"`
}

// CheckSaDevice 用于处理检查SA设备绑定情况接口的请求
func CheckSaDevice(c *gin.Context) {
	var (
		err  error
		resp CheckSaDeviceResp
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()
	resp.Revision = os.Getenv("GIT_COMMIT")
	resp.Version = types.Version
	resp.MinVersion = types.MinVersion
	if _, err = entity.GetSaDevice(); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.IsBind = true
	return
}

type IsAccessAllowResp struct {
	AccessAllow bool `json:"access_allow"` // 是否允许访问(判断用户token是否在该SA中有效)
}

// IsAccessAllow 是否能访问该SA
func IsAccessAllow(c *gin.Context) {
	var (
		err  error
		resp IsAccessAllowResp
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()
	resp.AccessAllow = session.Get(c) != nil
}
