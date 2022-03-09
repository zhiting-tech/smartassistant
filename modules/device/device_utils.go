package device

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/event"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
	plugin2 "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
	"gorm.io/gorm"
)

// IsPermit 判断用户是否有权限
func IsPermit(c *gin.Context, p types.Permission) bool {
	u := session.Get(c)
	return u != nil && entity.JudgePermit(u.UserID, p)
}

// ControlPermissions 根据配置获取设备所有控制权限
func ControlPermissions(d entity.Device) ([]types.Permission, error) {
	as, err := GetControlAttributes(d)
	if err != nil {
		logger.Error("GetControlAttributesErr", err)
		return nil, err
	}

	target := types.DeviceTarget(d.ID)
	res := make([]types.Permission, 0)
	for _, attr := range as {
		// 属性的状态需要是可控制可读，才需要被权限控制
		if attr.Attribute.Permission != attribute.AttrPermissionALLControl {
			continue
		}
		name := attr.Attribute.Attribute
		attr := entity.PluginDeviceAttr(attr.InstanceID, attr.Attribute.Attribute)
		p := types.Permission{
			Name:      name,
			Action:    "control",
			Target:    target,
			Attribute: attr,
		}
		res = append(res, p)
	}
	return res, nil
}

// Permissions 根据配置获取设备所有权限
func Permissions(d entity.Device) (ps []types.Permission, err error) {
	ps = append(ps, ManagePermissions(d)...)
	ps = append(ps, types.NewDeviceUpdate(d.ID))

	if d.Model == types.SaModel {
		return
	}

	controlPermission, err := ControlPermissions(d)
	if err != nil {
		return
	}

	// 非SA设备可配置删除设备权限,控制设备权限
	ps = append(ps, controlPermission...)
	ps = append(ps, types.NewDeviceDelete(d.ID))
	return
}

// IsDeviceControlPermit 控制设备的websocket命令 是否有权限
func IsDeviceControlPermit(areaID uint64, userID int, pluginID, identity string, data json.RawMessage) bool {
	d, err := entity.GetPluginDevice(areaID, pluginID, identity)
	if err != nil {
		err = errors.New(status.DeviceNotExist)
		logger.Warning(err)
		return false
	}

	var req plugin2.SetRequest
	if err = json.Unmarshal(data, &req); err != nil {
		logrus.Errorf("IsDeviceControlPermit unmarshal err: %s", err.Error())
		return false
	}
	up, err := entity.GetUserPermissions(userID)
	if err != nil {
		return false
	}
	for _, attr := range req.Attributes {
		logger.Debug(d, attr)
		if !up.IsDeviceAttrControlPermit(d.ID, attr.InstanceID, attr.Attribute) {
			return false
		}
	}
	return true
}

// ManagePermissions 设备的管理权限
func ManagePermissions(d entity.Device) []types.Permission {
	var permissions = make([]types.Permission, 0)
	// TODO 设备的固件升级功能是否能和设备的其他控制属性一样从插件获取？
	if d.Model == types.SaModel {
		permissions = append(permissions, types.NewDeviceManage(d.ID, "固件升级", types.FwUpgrade))
		permissions = append(permissions, types.NewDeviceManage(d.ID, "软件升级", types.SoftwareUpgrade))
	}
	return permissions
}

// BatchInsertChildDevice 通过das批量插入子设备
func BatchInsertChildDevice(d *entity.Device, das plugin.DeviceInstances, tx *gorm.DB) error {
	var childDevices []*entity.Device
	for _, ins := range das.Instances {
		if ins.Attributes[0].Attribute.Attribute != "is_child_instance" {
			continue
		}

		var childDevice entity.Device
		for _, attr := range ins.Attributes {
			switch attr.Attribute.Attribute {
			case "name":
				childDevice.Name = fmt.Sprintf("%s", attr.Val)
			case "manufacturer":
				childDevice.Manufacturer = fmt.Sprintf("%s", attr.Val)
			case "model":
				childDevice.Model = fmt.Sprintf("%s", attr.Val)
			}
		}
		childDevice.PID = d.ID // 所属父设备

		// 三个字段为唯一索引
		childDevice.AreaID = d.AreaID                                                       // 家庭ID
		childDevice.LocationID = d.LocationID                                               // 房间ID
		childDevice.PluginID = d.PluginID                                                   // 插件标识
		childDevice.Identity = fmt.Sprintf("childDevice-%s-%d", d.Identity, ins.InstanceId) // 使用Identity-instance_id作为新的Identity

		// 子设备的thing_model字段，需要从父设备中获取
		var childDeviceDas plugin.DeviceInstances
		childDeviceDas.Identity = childDevice.Identity
		childDeviceDas.OTASupport = das.OTASupport
		childDeviceDas.Instances = das.Instances[ins.InstanceId-1 : ins.InstanceId]
		childDevice.ThingModel, _ = json.Marshal(childDeviceDas)
		// 设置影子模型
		shadow := entity.NewShadow()
		for _, ins := range childDeviceDas.Instances {
			for _, attr := range ins.Attributes {
				shadow.UpdateReported(ins.InstanceId, attr.Attribute)
			}
		}
		childDevice.Shadow, _ = json.Marshal(shadow)

		childDevices = append(childDevices, &childDevice)
	}
	if len(childDevices) > 0 {
		if err := entity.BatchAddChildDevice(childDevices, tx); err != nil {
			return err
		}
		for _, childDevice := range childDevices {
			// 为子设备添加权限
			if err := AddDevicePermissionForRoles(*childDevice, tx); err != nil {
				return err
			}
		}
		event.GetServer().Notify(event.NewEventMessage(event.DeviceIncrease, d.AreaID))
	}
	return nil
}
