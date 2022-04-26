package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	url2 "net/url"
	"strings"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

// RemoveDevice 删除设备,断开相关连接和回收资源
func RemoveDevice(ctx context.Context, deviceID int) (err error) {
	d, err := entity.GetDeviceByID(deviceID)
	if err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}

	if d.Model == types.SaModel {
		return errors.New(status.ForbiddenBindOtherSA)
	}
	identify := Identify{
		PluginID: d.PluginID,
		IID:      d.IID,
		AreaID:   d.AreaID,
	}
	if err = DisconnectDevice(ctx, identify, nil); err != nil {
		logger.Error("disconnect err:", err)
	}

	if err = entity.DelDeviceByID(deviceID); err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	return
}

// SetAttributes 通过插件设置设备的属性
func SetAttributes(ctx context.Context, pluginID string, areaID uint64, setReq sdk.SetRequest) (err error) {
	_, err = GetGlobalClient().SetAttributes(ctx, pluginID, areaID, setReq)
	return
}

// OTA 更新插件的设备的固件
func OTA(ctx context.Context, areaID uint64, pluginID, iid, firmwareURL string) (err error) {
	d, err := entity.GetPluginDevice(areaID, pluginID, iid)
	if err != nil {
		return
	}
	identify := Identify{
		PluginID: d.PluginID,
		IID:      d.IID,
		AreaID:   d.AreaID,
	}
	return GetGlobalClient().OTA(ctx, identify, firmwareURL)
}

func ThingModelToEntity(iid string, tm thingmodel.ThingModel, pluginID string, areaID uint64) (d entity.Device, err error) {
	info, err := tm.GetInfo(iid)
	tmJson, err := json.Marshal(tm)
	if err != nil {
		return
	}
	conf := GetGlobalClient().DeviceConfig(pluginID, info.Model)
	name := conf.Name
	if conf.Name == "" {
		name = info.Model
	}
	d = entity.Device{
		Name:         name,
		Model:        info.Model,
		Manufacturer: info.Manufacturer,
		IID:          info.IID,
		PluginID:     pluginID,
		AreaID:       areaID,
		ThingModel:   tmJson,
	}
	shadow := entity.NewShadow()
	for _, instance := range tm.Instances {
		for _, srv := range instance.Services {
			for _, attr := range srv.Attributes {
				shadow.UpdateReported(instance.IID, attr.AID, attr.Val)
			}
		}
	}
	d.Shadow, err = json.Marshal(shadow)
	if err != nil {
		return
	}
	d.Type = conf.Type.String()
	return
}
func InstanceToEntity(instance thingmodel.Instance, pluginID string, areaID uint64) (d entity.Device, err error) {
	info, err := instance.GetInfo()
	if err != nil {
		return
	}
	tm := thingmodel.ThingModel{
		Instances:  []thingmodel.Instance{instance},
		OTASupport: false,
	}
	tmJson, err := json.Marshal(tm)
	if err != nil {
		return
	}
	conf := GetGlobalClient().DeviceConfig(pluginID, info.Model)
	name := conf.Name
	if conf.Name == "" {
		name = info.Model
	}
	d = entity.Device{
		Name:         name,
		Model:        info.Model,
		Manufacturer: info.Manufacturer,
		IID:          info.IID,
		PluginID:     pluginID,
		AreaID:       areaID,
		ThingModel:   tmJson,
	}
	shadow := entity.NewShadow()
	for _, srv := range instance.Services {
		for _, attr := range srv.Attributes {
			shadow.UpdateReported(instance.IID, attr.AID, attr.Val)
		}
	}
	d.Shadow, err = json.Marshal(shadow)
	if err != nil {
		return
	}
	d.Type = conf.Type.String()
	return
}

func ConnectDevice(ctx context.Context, identify Identify, authParams map[string]interface{}) (das thingmodel.ThingModel, err error) {
	return GetGlobalClient().Connect(ctx, identify, authParams)
}

func DisconnectDevice(ctx context.Context, identify Identify, authParams map[string]interface{}) error {
	return GetGlobalClient().Disconnect(ctx, identify, authParams)
}

func ConcatPluginPath(pluginID string, paths ...string) string {
	paths = append([]string{"plugin", pluginID}, paths...)
	return url.ConcatPath(paths...)
}

// DeviceLogoURL 设备Logo图片地址
func DeviceLogoURL(req *http.Request, pluginID, model string) string {
	logo := GetGlobalClient().DeviceConfig(pluginID, model).Logo
	return PluginTargetURL(req, pluginID, model, logo)
}

func PluginTargetURL(req *http.Request, pluginID, model, target string) string {
	if model == types.SaModel {
		return url.SAImageUrl(req)
	}
	path := ConcatPluginPath(pluginID, target)
	return url.BuildURL(path, nil, req).String()
}

type URL struct {
	url *url2.URL
}

// String 插件URL
func (u URL) String() string {
	return u.url.String()
}

// PluginPath 插件相对路径
func (u URL) PluginPath() string {
	uri := u.url.RequestURI()
	strs := strings.SplitN(uri, "/", 3)
	if len(strs) >= 3 {
		return strs[2]
	}
	return ""
}

// ControlURL 返回设备的插件控制页url
func ControlURL(d entity.Device, req *http.Request, userID int) (*URL, error) {
	if d.Model == types.SaModel {
		return nil, nil
	}
	token, err := oauth.GetUserPluginToken(userID, req, d.AreaID)
	if err != nil {
		return nil, err
	}

	q := map[string]interface{}{
		"device_id": d.ID,
		"iid":       d.IID,
		"model":     d.Model,
		"name":      d.Name,
		"token":     token,
		"type":      GetGlobalClient().DeviceConfig(d.PluginID, d.Model).Type,
		"sa_id":     config.GetConf().SmartAssistant.ID,
		"plugin_id": d.PluginID,
	}
	controlPath := ConcatPluginPath(d.PluginID, GetGlobalClient().DeviceConfig(d.PluginID, d.Model).Control)
	return &URL{url.BuildURL(controlPath, q, req)}, nil
}

// ArchiveURL 插件的前端压缩包地址
func ArchiveURL(pluginID string, req *http.Request) string {

	fileName := fmt.Sprintf("%s.zip", pluginID)
	path := ConcatPluginPath(pluginID, "resources/archive", fileName)
	return url.BuildURL(path, nil, req).String()
}
