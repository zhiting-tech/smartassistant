package clouddisk

import (
	"bytes"
	"context"
	"encoding/json"
	errors2 "errors"
	"fmt"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/zhiting-tech/smartassistant/pkg/http/httpclient"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/extension"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
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
	if !extension.HasExtensionWithContext(c.Request.Context(), types.CloudDisk) {
		err = errors.New(errors.BadRequest)
		return
	}

	accessToken := c.GetHeader(types.SATokenKey)
	ctx := c.Request.Context()
	if result, err = DelCloudDiskWithContext(c.Request.Context(), accessToken, isDelCloudDiskFile, areaID); err != nil {
		return
	}

	if !isDelCloudDiskFile && result.Data.RemoveStatus != CloudDiskDelSuccess {
		err = DelAreaWithContext(c.Request.Context(), areaID)
		result.Data.RemoveStatus = CloudDiskDelSuccess
		return
	}

	if result.Data.RemoveStatus != CloudDiskDelSuccess {
		_, loaded := onDelAreas.LoadOrStore(areaID, nil)
		if !loaded {
			go beginPollingDelCloudDisk(ctx, accessToken, isDelCloudDiskFile, areaID)
		}
	}
	return
}

// DelCloudDiskWithContext 删除网盘资源
func DelCloudDiskWithContext(ctx context.Context, accessToken string, isDelCloudDiskFile bool, areaID uint64) (result DelCloudDiskResp, err error) {
	url := fmt.Sprintf("http://%s/wangpan/api/folders", types.CloudDiskAddr)
	param := map[string]interface{}{
		"is_del_cloud_disk": isDelCloudDiskFile,
	}

	// 2  序列化数据
	content, _ := json.Marshal(param)
	// 复制父span到新的context中, 保证context被返回还保留着spanID的关联
	newCtx := trace.ContextWithSpan(context.Background(), trace.SpanFromContext(ctx))
	request, err := http.NewRequestWithContext(newCtx, http.MethodDelete, url, bytes.NewReader(content))
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	// 3、获取token并放入header
	request.Header.Set(types.SATokenKey, accessToken)
	client := httpclient.NewHttpClient(httpclient.WithTimeout(60 * time.Second))
	response, err := client.Do(request)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if response.StatusCode != http.StatusOK {
		logger.Errorf("del cloud disk err of response statusCode %d", response.StatusCode)
		err = errors.Wrap(errors2.New("request cloud disk err"), errors.InternalServerErr)
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
		err = DelAreaWithContext(newCtx, areaID)
	}
	return
}

// beginPollingDelCloudDisk 开始轮询网盘删除
func beginPollingDelCloudDisk(ctx context.Context, accessToken string, isDelCloudDiskFile bool, areaID uint64) {
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
			result, err = DelCloudDiskWithContext(ctx, accessToken, isDelCloudDiskFile, areaID)
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

func DelAreaWithContext(ctx context.Context, areaID uint64) (err error) {
	_, err = entity.GetAreaByID(areaID)
	if err != nil {
		return
	}
	plugins, err := entity.GetInstalledPlugins()
	if err != nil {
		return
	}

	if err = cloud.RemoveSAWithContext(ctx, areaID); err != nil {
		return
	}

	if err = entity.DelAreaByID(areaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

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
