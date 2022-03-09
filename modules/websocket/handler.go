package websocket

import (
	"encoding/json"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	version2 "github.com/zhiting-tech/smartassistant/modules/utils/version"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

func GetAttrs(cs callService) (result Result, err error) {
	result = make(Result)
	user := cs.callUser
	d, err := device.GetInstances(user.AreaID, user.UserID, cs.Domain, cs.Identity)
	if err != nil {
		return
	}
	result["device"] = d
	return
}

func SetAttrs(cs callService) (result Result, err error) {

	result = make(Result)
	user := cs.callUser
	_, err = entity.GetDeviceByIdentity(cs.Identity)
	if err == nil {
		// 根据插件配置判断用户是否具有权限
		if !device.IsDeviceControlPermit(user.AreaID, user.UserID, cs.Domain, cs.Identity, cs.ServiceData) {
			err = errors.New(status.Deny)
			return
		}
	} else {
		if err != gorm.ErrRecordNotFound {
			return
		}
		err = nil
	}

	err = plugin.SetAttributes(user.AreaID, cs.Domain, cs.Identity, cs.ServiceData)
	if err != nil {
		return
	}
	return
}

// ConnectDevice 连接设备 TODO 直接替代添加设备接口？
func ConnectDevice(cs callService) (result Result, err error) {
	result = make(Result)
	var authParams map[string]string
	if err = json.Unmarshal(cs.ServiceData, &authParams); err != nil {
		return
	}
	d, err := plugin.ConnectDevice(cs.Identity, cs.Domain, authParams)
	if err != nil {
		return
	}
	result["device"] = d

	// 自动加入设备列表
	// deviceEntity, err := plugin.GetInfoFromDeviceAttrs(cs.UID, d)
	// if err != nil {
	// 	return
	// }
	// if err = device.Create(cs.callUser.AreaID, &deviceEntity); err != nil {
	// 	return
	// }

	return
}

// DisconnectDevice 设备断开连接（取消配对等） TODO 直接替代删除设备接口？
func DisconnectDevice(cs callService) (result Result, err error) {

	result = make(Result)
	var authParams map[string]string
	if err = json.Unmarshal(cs.ServiceData, &authParams); err != nil {
		return
	}
	err = plugin.DisconnectDevice(cs.Identity, cs.Domain, authParams)
	if err != nil {
		return
	}
	return
}

// UpdateThingModel 更新设备的影子模型字段
func UpdateThingModel(cs callService) (result Result, err error) {
	result = make(Result)
	user := cs.callUser

	d, err := entity.GetPluginDevice(user.AreaID, cs.Domain, cs.Identity)
	if err != nil {
		return
	}
	// 获取最新的das
	das, err := plugin.GetGlobalClient().GetAttributes(d)
	if err != nil {
		return
	}
	// 转换成json格式
	thingModel, err := json.Marshal(das)
	if err != nil {
		return
	}
	// 新建设备影子并更新
	shadow := entity.NewShadow()
	for _, ins := range das.Instances {
		for _, attr := range ins.Attributes {
			shadow.UpdateReported(ins.InstanceId, attr.Attribute)
		}
	}
	shadowStr, err := json.Marshal(shadow)
	if err != nil {
		return
	}
	// 更新对应的字段
	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {
		// 更新影子模型字段
		deviceValues := map[string]interface{}{"thing_model": thingModel, "shadow": shadowStr}
		if err = entity.UpdateDeviceById(d.ID, deviceValues, tx); err != nil {
			return err
		}
		// 先删除子设备，再重新添加
		if err = entity.DelDeviceByPID(d.ID, tx); err != nil {
			return err
		}
		// 批量插入子设备
		if err = device.BatchInsertChildDevice(&d, das, tx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return
	}

	// 返回最新的数据
	newDevice, err := device.GetInstances(user.AreaID, user.UserID, cs.Domain, cs.Identity)
	if err != nil {
		return
	}
	result["device"] = newDevice

	return
}

// CheckUpdate 检查设备是否有更新
func CheckUpdate(cs callService) (result Result, err error) {
	result = make(Result)
	user := cs.callUser
	d, err := device.GetInstances(user.AreaID, user.UserID, cs.Domain, cs.Identity)
	if err != nil {
		return
	}

	info, err := d.GetInfo()
	if err != nil {
		return
	}

	latestFireware, err := cloud.GetLatestFirmware(cs.Domain, info.Model)
	if err != nil {
		return
	}
	ok, err := version2.Greater(latestFireware.Version, info.Version)
	if err != nil {
		return
	}

	result["update_available"] = ok
	result["current_version"] = info.Version
	result["latest_firmware"] = latestFireware
	return
}

// OTA 更新固件
func OTA(cs callService) (result Result, err error) {

	user := cs.callUser

	d, err := entity.GetPluginDevice(user.AreaID, cs.Domain, cs.Identity)
	if err != nil {
		return
	}

	instances, err := device.GetUserDeviceInstances(user.UserID, d)
	if err != nil {
		return
	}

	info, err := instances.GetInfo()
	if err != nil {
		return
	}
	firmware, err := cloud.GetLatestFirmware(cs.Domain, info.Model)
	if err != nil {
		return
	}
	result = make(Result)
	err = plugin.OTA(user.AreaID, cs.Domain, cs.Identity, firmware.URL)
	if err != nil {
		return
	}

	// 硬件响应ota成功仅表示固件flash成功，真正是否成功需要等待硬件重启后才能确定固件版本号
	// FIXME 暂时先清空，等待下次请求获取设备信息时重新请求
	if err = d.Clear(); err != nil {
		return
	}
	return
}

func RegisterCmd() {
	RegisterCallFunc(serviceConnect, ConnectDevice)
	RegisterCallFunc(serviceDisconnect, DisconnectDevice)
	RegisterCallFunc(serviceSetAttributes, SetAttrs)
	RegisterCallFunc(serviceGetAttributes, GetAttrs)
	RegisterCallFunc(serviceUpdateThingModel, UpdateThingModel)
	RegisterCallFunc(serviceOTA, OTA)
	RegisterCallFunc(serviceCheckUpdate, CheckUpdate)
}
