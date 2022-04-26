package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/modules/supervisor/proto"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type GetSystemInfoResp struct {
	Version string `json:"version"`
}

func GetSystemInfo(c *gin.Context) {

	var (
		resp     GetSystemInfoResp
		grpcResp *proto.GetSystemInfoResp
		err      error
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	grpcResp, err = supervisor.GetClient().SystemInfoWithContext(c.Request.Context())
	if err != nil {
		err = errors.Wrap(err, status.GetFirmwareVersionErr)
		return
	}
	resp.Version = grpcResp.Version

}

type GetSystemLastVersionResp struct {
	LatestVersion string `json:"latest_version"`
}

func GetSystemLastVersion(c *gin.Context) {
	var (
		resp   GetSystemLastVersionResp
		err    error
		result *cloud.FirmwareLastVersionHttpResult
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	result, err = cloud.GetLastFirmwareVersionWithContext(c.Request.Context())
	if err != nil {
		err = errors.Wrap(err, status.GetFirmwareVersionErr)
		return
	}

	if result.Data.Version == "" {
		err = errors.Wrap(err, status.GetFirmwareVersionErr)
		return
	}
	resp.LatestVersion = result.Data.Version
}
