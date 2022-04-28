package clouddisk

import (
	"context"
	"encoding/json"
	errors2 "errors"
	"fmt"
	"github.com/zhiting-tech/smartassistant/modules/types"
	url2 "github.com/zhiting-tech/smartassistant/modules/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

// HandlerCloudDiskRequest 处理网盘请求
func HandlerCloudDiskRequest(request *http.Request, accessToken string) (resp []byte, err error) {
	request.Header.Set(types.SATokenKey, accessToken)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if response.StatusCode != http.StatusOK {
		err = errors.Wrap(errors2.New("request cloud disk err"), errors.InternalServerErr)
		return
	}
	defer response.Body.Close()
	resp, _ = ioutil.ReadAll(response.Body)
	return
}

type GetPartitionInfoResp struct {
	Status int           `json:"status"`
	Reason string        `json:"reason"`
	Data   PartitionInfo `json:"data"`
}

type PartitionInfo struct {
	FolderInfo []FolderInfo `json:"folder_info"`
}

type FolderInfo struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	PartitionPath string `json:"partition_path"`
}

// GetCloudDiskPartitionInfo 获取网盘指定分区能看到的文件夹
func GetCloudDiskPartitionInfo(accessToken, path string, ctx context.Context) (result GetPartitionInfoResp, err error) {
	parseURL, err := url.Parse(fmt.Sprintf("http://%s/wangpan/api/common/partitions", types.CloudDiskAddr))
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	parseURL.RawQuery = url2.Join(url2.BuildQuery(map[string]interface{}{
		"path": path,
	}))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parseURL.String(), nil)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp, err := HandlerCloudDiskRequest(request, accessToken)
	err = json.Unmarshal(resp, &result)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if result.Status != 0 {
		err = errors2.New(result.Reason)
		return
	}
	return
}

type GetPoolListResp struct {
	Status int      `json:"status"`
	Reason string   `json:"reason"`
	Data   ListPool `json:"data"`
}

type ListPool struct {
	Pools []PoolInfo `json:"list"`
}

type PoolInfo struct {
	Name string `json:"name"` // 存储池名称
}

// GetCloudDiskPools 获取网盘存储池列表
func GetCloudDiskPools(accessToken string, ctx context.Context) (result GetPoolListResp, err error) {
	reqUrl := fmt.Sprintf("http://%s/wangpan/api/pools", types.CloudDiskAddr)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp, err := HandlerCloudDiskRequest(request, accessToken)
	err = json.Unmarshal(resp, &result)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if result.Status != 0 {
		err = errors2.New(result.Reason)
		return
	}
	return
}

type GetPoolInfoResp struct {
	Status int      `json:"status"`
	Reason string   `json:"reason"`
	Data   InfoResp `json:"data"`
}

type InfoResp struct {
	Name string           `json:"name"` // 存储池名称
	Lv   []*LogicalVolume `json:"lv"`   // 逻辑分区
}

// LogicalVolume 逻辑分区
type LogicalVolume struct {
	Id   string `json:"id"`   // 存储池唯一标识符
	Name string `json:"name"` // 存储池名称
}

// GetCloudDiskPoolInfo 获取网盘存储池详情
func GetCloudDiskPoolInfo(accessToken, poolName string, ctx context.Context) (result GetPoolInfoResp, err error) {
	reqUrl := fmt.Sprintf("http://%s/wangpan/api/pools/%s", types.CloudDiskAddr, poolName)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp, err := HandlerCloudDiskRequest(request, accessToken)
	err = json.Unmarshal(resp, &result)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if result.Status != 0 {
		err = errors2.New(result.Reason)
		return
	}
	return
}

type GetFolderInfoResp struct {
	Status int        `json:"status"`
	Reason string     `json:"reason"`
	Data   ListFolder `json:"data"`
}

type ListFolder struct {
	FolderList []Info `json:"list"`
}

type Info struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
	Path string `json:"path"`
}

// GetCloudDiskFolderInfo 获取网盘文件详情
func GetCloudDiskFolderInfo(accessToken, path string, ctx context.Context) (result GetFolderInfoResp, err error) {
	parseURL, err := url.Parse(fmt.Sprintf("http://%s/wangpan/api/resources/%s", types.CloudDiskAddr, path))
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	parseURL.RawQuery = url2.Join(url2.BuildQuery(map[string]interface{}{
		"type":      0,
		"page":      0,
		"page_size": 0,
	}))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parseURL.String(), nil)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp, err := HandlerCloudDiskRequest(request, accessToken)
	err = json.Unmarshal(resp, &result)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if result.Status != 0 {
		err = errors2.New(result.Reason)
		return
	}
	return
}

type GetResourcePathResp struct {
	Status int          `json:"status"`
	Reason string       `json:"reason"`
	Data   resourcePath `json:"data"`
}

type resourcePath struct {
	Path     string `json:"path"`
	ShowPath string `json:"show_path"`
}

// GetCloudDiskResourcePath 获取网盘文件路径
func GetCloudDiskResourcePath(accessToken, path string, ctx context.Context) (result GetResourcePathResp, err error) {
	parseURL, err := url.Parse(fmt.Sprintf("http://%s/wangpan/api/common/resource/path", types.CloudDiskAddr))
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	parseURL.RawQuery = url2.Join(url2.BuildQuery(map[string]interface{}{
		"path": path,
	}))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parseURL.String(), nil)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp, err := HandlerCloudDiskRequest(request, accessToken)
	err = json.Unmarshal(resp, &result)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if result.Status != 0 {
		err = errors2.New(result.Reason)
		return
	}
	return
}
