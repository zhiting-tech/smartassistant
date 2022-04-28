package supervisor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

const (
	EnvironmentSoftwareVersionKey = "SOFTWARE_SERVICE_VERSION"
)

type updateInfoResp struct {
	Version string `json:"version"`
}

// UpdateInfo 查看更新信息
func UpdateInfo(c *gin.Context) {
	var (
		resp updateInfoResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	resp.Version, err = getLocalSoftwareVersion()
	if err != nil {
		errors.Wrap(err, status.GetImageVersionErr)
		return
	}

}

func getLocalSoftwareVersion() (version string, err error) {
	var (
		data   []byte
		envMap map[string]string
		ok     bool
	)

	if data, err = os.ReadFile(filepath.Join(config.GetConf().SmartAssistant.RuntimePath, ".env")); err != nil {
		return
	}
	if envMap, err = godotenv.Unmarshal(string(data)); err != nil {
		return
	}

	if version, ok = envMap[EnvironmentSoftwareVersionKey]; !ok || version == "" {
		err = fmt.Errorf("env %s is empty", EnvironmentSoftwareVersionKey)
		return
	}

	return
}

type UpdateLastVersionResp struct {
	LatestVersion string `json:"latest_version"`
}

func UpdateLastVersion(c *gin.Context) {
	var (
		resp   UpdateLastVersionResp
		err    error
		result *cloud.SoftwareLastVersionHttpResult
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if result, err = cloud.GetLastSoftwareVersionWithContext(c.Request.Context()); err != nil {
		err = errors.Wrap(err, status.GetImageVersionErr)
		return
	}

	if result.Data.Version == "" {
		err = errors.New(status.GetImageVersionErr)
		return
	}

	resp.LatestVersion = result.Data.Version
}
