package cloud

import (
	"context"
	errors2 "errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/http/httpclient"
	"github.com/zhiting-tech/smartassistant/pkg/logger"

	jsoniter "github.com/json-iterator/go"
)

var ErrNoFirmware = errors2.New("no firmware found")

type DeviceType string
type DeviceSubType string

const (
	TypeLight          DeviceType = "light"           // 灯
	TypeSwitch         DeviceType = "switch"          // 开关
	TypeOutlet         DeviceType = "outlet"          // 插座
	TypeRoutingGateway DeviceType = "routing_gateway" // 路由网关
	TypeSecurity       DeviceType = "security"        // 安防
)

type Brand struct {
	LogoURL      string `json:"logo_url"`
	Name         string `json:"name"`
	PluginAmount int    `json:"plugin_amount"` // 插件数量
}

type Plugin struct {
	ID      int    `json:"id"`
	UID     string `json:"uid"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Version string `json:"version"`
	Brand   string `json:"brand"`
	Intro   string `json:"intro"`
}

type Device struct {
	Model string     `json:"model" `
	Name  string     `json:"name"`
	Type  DeviceType `json:"type"` // 设备类型

	Logo         string `json:"logo" `        // 设备logo相对路径
	Control      string `json:"control"`      // 设备控制页面相对路径
	Provisioning string `json:"provisioning"` // 设备置网页面相对路径
}

func GetBrandsWithContext(ctx context.Context) (brands []Brand, err error) {

	url := fmt.Sprintf("%s/common/brands", config.GetConf().SmartCloud.URL())
	logger.Debug(url)

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = httpclient.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get %s fail, status code: %d", url, resp.StatusCode)
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	status := jsoniter.Get(data, "status").ToInt()
	if status != 0 {
		err = fmt.Errorf("invalid response %s", string(data))
		return
	}

	any := jsoniter.Get(data, "data", "list")
	any.ToVal(&brands)
	err = any.LastError()
	return
}

func GetBrandsMapWithContext(ctx context.Context) (brandsMap map[string]Brand, err error) {
	brands, err := GetBrandsWithContext(ctx)
	if err != nil {
		return
	}
	brandsMap = make(map[string]Brand)
	for _, brand := range brands {
		brandsMap[brand.Name] = brand
	}
	return
}

type BrandInfo struct {
	Brand
	Plugins []Plugin `json:"plugins"`
}

func GetBrandInfoWithContext(ctx context.Context, brandName string) (brand BrandInfo, err error) {
	url := fmt.Sprintf("%s/common/brands/name/%s", config.GetConf().SmartCloud.URL(), brandName)
	logger.Debug(url)

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = httpclient.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get %s fail, status code: %d", url, resp.StatusCode)
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	status := jsoniter.Get(data, "status").ToInt()
	if status != 0 {
		err = fmt.Errorf("invalid response %s", string(data))
		return
	}

	any := jsoniter.Get(data, "data")
	logger.Debug(any.ToString())
	any.ToVal(&brand)
	err = any.LastError()
	return
}

func GetPluginsWithContext(ctx context.Context) (plugins []Plugin, err error) {
	url := fmt.Sprintf("%s/common/plugins", config.GetConf().SmartCloud.URL())
	logger.Debug(url)

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = httpclient.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get %s fail, status code: %d", url, resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	status := jsoniter.Get(data, "status").ToInt()
	if status != 0 {
		err = fmt.Errorf("invalid response %s", string(data))
		return
	}

	any := jsoniter.Get(data, "data", "list")
	any.ToVal(&plugins)
	err = any.LastError()
	return
}

func GetPluginWithContext(ctx context.Context, PluginUID string) (plugin Plugin, err error) {
	url := fmt.Sprintf("%s/common/plugins/uid/%s",
		config.GetConf().SmartCloud.URL(), PluginUID)
	logger.Debug(url)

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = httpclient.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get %s fail, status code: %d", url, resp.StatusCode)
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	status := jsoniter.Get(data, "status").ToInt()
	if status != 0 {
		err = fmt.Errorf("invalid response %s", string(data))
		return
	}

	any := jsoniter.Get(data, "data", "plugin")
	any.ToVal(&plugin)
	err = any.LastError()
	return
}

type Firmware struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	Info    string `json:"info"`
}

// GetLatestFirmwareWithContext 获取最新的固件
func GetLatestFirmwareWithContext(ctx context.Context, pluginID, model string) (firmware Firmware, err error) {

	url := fmt.Sprintf("%s/common/plugins/uid/%s/model/%s/firmwares",
		config.GetConf().SmartCloud.URL(), pluginID, model)
	logger.Debug(url)

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = httpclient.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get %s fail, status code: %d", url, resp.StatusCode)
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	status := jsoniter.Get(data, "status").ToInt()
	if status != 0 {
		err = fmt.Errorf("invalid response %s", string(data))
		return
	}

	var firmwares []Firmware
	any := jsoniter.Get(data, "data", "firmwares")
	any.ToVal(&firmwares)
	err = any.LastError()
	if len(firmwares) != 0 {
		return firmwares[0], nil
	}
	err = ErrNoFirmware
	return
}

type App struct {
	AppID  int    `json:"app_id"`
	Name   string `json:"name"`
	IsBind bool   `json:"is_bind"`
	Img    string `json:"img"`
	Link   string `json:"link"`
}

// GetAppList 获取第三方平台列表
func GetAppList(ctx context.Context, areaID uint64) (apps []App, err error) {
	path := fmt.Sprintf("apps?area_id=%d", areaID)
	resp, err := cloud.DoWithContext(ctx, path, http.MethodGet, nil)
	if err != nil {
		return
	}

	any := jsoniter.Get(resp, "data", "apps")
	any.ToVal(&apps)
	err = any.LastError()

	return
}

// UnbindApp 解绑第三方平台
func UnbindApp(ctx context.Context, areaID uint64, appID int) (err error) {
	path := fmt.Sprintf("apps/%d/areas/%d", appID, areaID)
	_, err = cloud.DoWithContext(ctx, path, http.MethodDelete, nil)

	return
}
