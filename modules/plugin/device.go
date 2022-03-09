package plugin

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// RemoveDevice 删除设备,断开相关连接和回收资源
func RemoveDevice(deviceID int) (err error) {
	d, err := entity.GetDeviceByID(deviceID)
	if err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}

	if d.Model == types.SaModel {
		return errors.New(status.ForbiddenBindOtherSA)
	}

	if err = DisconnectDevice(d.Identity, d.PluginID, nil); err != nil {
		logrus.Error("disconnect err:", err)
	}

	if err = entity.DelDeviceByID(deviceID); err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	if err = entity.DelDeviceByPID(deviceID, entity.GetDB()); err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	return
}

// SetAttributes 通过插件设置设备的属性
func SetAttributes(areaID uint64, pluginID, identity string, data json.RawMessage) (err error) {
	d, err := entity.GetPluginDevice(areaID, pluginID, identity)
	if err != nil {
		return
	}

	_, err = GetGlobalClient().SetAttributes(d, data)
	return
}

// OTA 更新插件的设备的固件
func OTA(areaID uint64, pluginID, identity, firmwareURL string) (err error) {
	d, err := entity.GetPluginDevice(areaID, pluginID, identity)
	if err != nil {
		return
	}

	return GetGlobalClient().OTA(d, firmwareURL)
}

// GetInstanceControlAttributes 获取实例的控制属性
func GetInstanceControlAttributes(instance Instance) (attributes []entity.Attribute) {
	for _, attr := range instance.Attributes {

		// 仅返回能控制的属性
		// TODO 这里不能只判断名称
		if attr.Attribute.Attribute == "name" {
			continue
		}

		a := entity.Attribute{
			Attribute:  attr.Attribute,
			InstanceID: instance.InstanceId,
		}
		attributes = append(attributes, a)
	}
	return
}

func ConnectDevice(identity, pluginID string, authParams map[string]string) (das DeviceInstances, err error) {
	return GetGlobalClient().Connect(identity, pluginID, authParams)
}

func DisconnectDevice(identity, pluginID string, authParams map[string]string) error {
	return GetGlobalClient().Disconnect(identity, pluginID, authParams)
}

func ConcatPluginPath(pluginID string, paths ...string) string {
	paths = append([]string{"api", "plugin", pluginID}, paths...)
	return url.ConcatPath(paths...)
}

// DeviceLogoURL 设备Logo图片地址
func DeviceLogoURL(req *http.Request, d entity.Device) string {
	return LogoURL(req, d.PluginID, d.Model, GetGlobalClient().DeviceConfig(d).Logo)
}

// LogoURL Logo图片地址
func LogoURL(req *http.Request, pluginID, model, logo string) string {
	if model == types.SaModel {
		return url.SAImageUrl(req)
	}
	path := ConcatPluginPath(pluginID, logo)
	return url.BuildURL(path, nil, req)
}

// PluginURL 返回设备的插件控制页url
func PluginURL(d entity.Device, req *http.Request, token string) string {
	if d.Model == types.SaModel {
		return ""
	}

	q := map[string]interface{}{
		"device_id": d.ID,
		"identity":  d.Identity,
		"model":     d.Model,
		"name":      d.Name,
		"token":     token,
		"type":      GetGlobalClient().DeviceConfig(d).Type,
		"sa_id":     config.GetConf().SmartAssistant.ID,
		"plugin_id": d.PluginID,
	}
	controlPath := ConcatPluginPath(d.PluginID, GetGlobalClient().DeviceConfig(d).Control)
	return url.BuildURL(controlPath, q, req)
}

// RelativeControlPath 返回设备的插件控制页相对路径
func RelativeControlPath(d entity.Device, token string) string {
	if d.Model == types.SaModel {
		return ""
	}

	q := map[string]interface{}{
		"device_id": d.ID,
		"identity":  d.Identity,
		"model":     d.Model,
		"name":      d.Name,
		"token":     token,
		"sa_id":     config.GetConf().SmartAssistant.ID,
		"plugin_id": d.PluginID,
	}
	return fmt.Sprintf("%s?%s", GetGlobalClient().DeviceConfig(d).Control, url.Join(url.BuildQuery(q)))
}

// ArchiveURL 插件的前端压缩包地址
func ArchiveURL(pluginID string, req *http.Request) string {

	fileName := fmt.Sprintf("%s.zip", pluginID)
	path := ConcatPluginPath(pluginID, "resources/archive", fileName)
	return url.BuildURL(path, nil, req)
}
