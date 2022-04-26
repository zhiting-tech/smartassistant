package supervisor

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

func getImagePath() string {
	imagePath := path.Join(config.GetConf().SmartAssistant.RuntimePath, "run", "supervisor", "images")
	if _, err := os.Stat(imagePath); err != nil {
		os.MkdirAll(imagePath, os.ModePerm)
	}
	return imagePath
}

type UpdateSystemReq struct {
	FileUrl   string `json:"-"`       // 固件升级包url
	Checksum  string `json:"-"`       // 固件升级包checksum
	Algorithm string `json:"-"`       // checksum算法
	Version   string `json:"version"` // 固件版本
}

func (req *UpdateSystemReq) Validate() error {
	if req.FileUrl == "" || req.Checksum == "" || req.Algorithm == "" {
		return errors.New(status.ParamRequireErr)
	}
	return nil
}

func (req *UpdateSystemReq) DownloadImage() (string, error) {
	// 下载文件
	uri, err := url.ParseRequestURI(req.FileUrl)
	if err != nil {
		return "", errors.Wrap(err, errors.InternalServerErr)
	}
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), path.Base(uri.Path))
	httpReq, err := http.NewRequest(http.MethodGet, req.FileUrl, nil)
	if err != nil {
		return "", errors.Wrap(err, errors.InternalServerErr)
	}
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", errors.Wrap(err, status.FirmwareDownloadErr)
	}
	defer func() {
		httpResp.Body.Close()
	}()

	// 保存文件同时计算checksum
	f, err := os.Create(path.Join(getImagePath(), filename))
	if err != nil {
		return "", errors.Wrap(err, errors.InternalServerErr)
	}
	defer f.Close()
	hash := sha256.New()
	mw := io.MultiWriter(f, hash)
	if _, err := io.Copy(mw, httpResp.Body); err != nil {
		return "", errors.Wrap(err, status.FirmwareDownloadErr)
	}
	sum := fmt.Sprintf("%x", hash.Sum(nil))

	// 校验checksum
	if sum != req.Checksum {
		return "", errors.New(status.ChecksumErr)
	}

	return filename, nil
}

func UpdateSystem(c *gin.Context) {
	var (
		req    UpdateSystemReq
		err    error
		image  string
		result *cloud.FirmwareLastVersionHttpResult
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		return
	}

	if result, err = cloud.GetLastFirmwareVersionWithContext(c.Request.Context()); err != nil {
		err = errors.Wrap(err, status.GetFirmwareVersionErr)
		return
	}

	req.FileUrl = result.Data.FileUrl
	req.Checksum = result.Data.Checksum
	req.Algorithm = result.Data.Algorithm
	if err = req.Validate(); err != nil {
		return
	}

	if image, err = req.DownloadImage(); err != nil {
		return
	}

	if err = supervisor.GetClient().UpdateSystemWithContext(c.Request.Context(), image); err != nil {
		return
	}

}
