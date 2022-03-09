package clouddisk

import (
	"bytes"
	"encoding/json"
	errors2 "errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/extension"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	CloudDiskOnDel = iota + 1
	CloudDiskDelErr
	CloudDiskDelSuccess
)

type DelCloudDiskResp struct {
	Status int           `json:"status"`
	Reason string        `json:"reason"`
	Data   DelAreaStatus `json:"data"`
}

type DelAreaStatus struct {
	RemoveStatus int `json:"remove_status"`
}

var onDelAreas sync.Map

// DelAreaCloudDisk 删除家庭的网盘资源
// isDelCloudDiskFile 是表示是否删除网盘文件，不管删不删除文件，一定会删除网盘记录
func DelAreaCloudDisk(c *gin.Context, isDelCloudDiskFile bool, areaID uint64) (result DelCloudDiskResp, err error) {
	if !extension.HasExtension(types.CloudDisk) {
		err = errors.New(errors.BadRequest)
		return
	}

	accessToken := c.GetHeader(types.SATokenKey)
	if result, err = DelCloudDisk(accessToken, isDelCloudDiskFile, areaID); err != nil {
		return
	}

	if !isDelCloudDiskFile && result.Data.RemoveStatus != CloudDiskDelSuccess {
		err = DelArea(areaID)
		result.Data.RemoveStatus = CloudDiskDelSuccess
		return
	}

	if result.Data.RemoveStatus != CloudDiskDelSuccess {
		_, loaded := onDelAreas.LoadOrStore(areaID, nil)
		if !loaded {
			go beginPollingDelCloudDisk(accessToken, isDelCloudDiskFile, areaID)
		}
	}
	return
}

// DelCloudDisk 删除网盘资源
func DelCloudDisk(accessToken string, isDelCloudDiskFile bool, areaID uint64) (result DelCloudDiskResp, err error) {
	url := fmt.Sprintf("http://%s/wangpan/api/folders", types.CloudDiskAddr)
	param := map[string]interface{}{
		"is_del_cloud_disk": isDelCloudDiskFile,
	}

	// 2  序列化数据
	content, _ := json.Marshal(param)
	request, err := http.NewRequest("DELETE", url, bytes.NewReader(content))
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	// 3、获取scope-token并放入header
	request.Header.Set("scope-token", accessToken)

	client := &http.Client{Timeout: 60 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if response.StatusCode != http.StatusOK {
		logger.Errorf("del cloud disk err of response statusCode %d", response.StatusCode)
		err = errors.New(errors.InternalServerErr)
		return
	}
	defer response.Body.Close()
	resp, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(resp, &result)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if result.Status != 0 {
		err = errors2.New(result.Reason)
		return
	}

	if result.Data.RemoveStatus == CloudDiskDelSuccess {
		err = DelArea(areaID)
	}
	return
}

// beginPollingDelCloudDisk 开始轮询网盘删除
func beginPollingDelCloudDisk(accessToken string, isDelCloudDiskFile bool, areaID uint64) {
	defer func() {
		onDelAreas.Delete(areaID)
	}()
	_, err := entity.GetAreaByID(areaID)
	if err != nil {
		logger.Errorf("pollingCloudDisk del err %s", err)
		return
	}
	logger.Info("begin polling del cloud disk")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			var result DelCloudDiskResp
			result, err = DelCloudDisk(accessToken, isDelCloudDiskFile, areaID)
			if err != nil {
				logger.Errorf("pollingCloudDisk del err %s", err.Error())
				return
			}
			if result.Data.RemoveStatus != CloudDiskDelSuccess {
				continue
			}
			return
		}
	}
}

func DelArea(areaID uint64) (err error) {
	_, err = entity.GetAreaByID(areaID)
	if err != nil {
		return
	}
	plugins, err := entity.GetInstalledPlugins()
	if err != nil {
		return
	}
	if err = entity.DelAreaByID(areaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	cloud.RemoveSA(areaID)
	for _, p := range plugins {
		if err2 := docker.GetClient().StopContainer(p.Image); err2 != nil {
			logger.Warnf("del area stop container %s", err2)
		}
		if err2 := docker.GetClient().RemoveContainer(p.Image); err2 != nil {
			logger.Warnf("del area remove container %s", err2)
		}
		if err2 := docker.GetClient().ImageRemove(p.Image); err2 != nil {
			logger.Warnf("del area remove image %s", err2)
		}
	}
	logger.Info("del area Success")
	return
}
