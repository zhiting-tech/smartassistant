package device

import (
	"encoding/json"
	"errors"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/event"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	errors2 "github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/utils"
	"gorm.io/gorm"
)

func Create(areaID uint64, device *entity.Device) (err error) {
	if device == nil {
		return errors.New("nil device")
	}

	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {

		device.AreaID = areaID
		// Create 添加设备
		switch device.Model {
		case types.SaModel:
			// 添加设备为SA时不需要添加设备影子
			if err = entity.AddSADevice(device, tx); err != nil {
				return err
			}
		default:
			if !plugin.GetGlobalClient().IsOnline(*device) {
				return errors2.New(status.DeviceOffline)
			}
			if err = entity.AddDevice(device, tx); err != nil {
				return err
			}
		}

		// 为所有角色增加改设备的权限
		return AddDevicePermissionForRoles(*device, tx)
	}); err != nil {
		return
	}
	return
}

// AddDevicePermissionForRoles 为所有角色增加设备权限
func AddDevicePermissionForRoles(device entity.Device, tx *gorm.DB) (err error) {

	// 将权限赋给给所有角色
	var roles []entity.Role
	// 使用同一个DB，保证在一个事务内
	roles, err = entity.GetRolesWithTx(tx, device.AreaID)
	if err != nil {
		return err
	}
	for _, role := range roles {
		// 查看角色设备权限模板配置
		if entity.IsDeviceActionPermit(role.ID, "manage", tx) {
			role.AddPermissionsWithDB(tx, ManagePermissions(device)...)
		}

		if entity.IsDeviceActionPermit(role.ID, "update", tx) {
			role.AddPermissionsWithDB(tx, types.NewDeviceUpdate(device.ID))
		}

		// SA设备不需要配置控制和删除权限
		if device.Model == types.SaModel {
			continue
		}
		if entity.IsDeviceActionPermit(role.ID, "control", tx) {
			var ps []types.Permission
			ps, err = ControlPermissions(device)
			if err != nil {
				logger.Error("ControlPermissionsErr:", err.Error())
				continue
			}
			role.AddPermissionsWithDB(tx, ps...)
		}

		if entity.IsDeviceActionPermit(role.ID, "delete", tx) {
			role.AddPermissionsWithDB(tx, types.NewDeviceDelete(device.ID))
		}
	}
	return
}

func GetShadow(d entity.Device) (shadow entity.Shadow, err error) {
	if d.IsInit() {
		// 从设备影子中获取属性
		shadow = entity.NewShadow()
		if err = json.Unmarshal(d.Shadow, &shadow); err != nil {
			return
		}
	}
	return
}

func GetThingModel(d entity.Device) (thingModel plugin.DeviceInstances, err error) {
	if d.IsInit() {
		// 从设备影子中获取属性
		if err = json.Unmarshal(d.ThingModel, &thingModel); err != nil {
			return
		}
	}
	return
}

// UpdateShadowReported 更新设备影子属性报告值
func UpdateShadowReported(d entity.Device, attr entity.Attribute) (err error) {
	// 从设备影子中获取属性
	shadow, err := GetShadow(d)
	if err != nil {
		return
	}
	shadow.UpdateReported(attr.InstanceID, attr.Attribute)
	d.Shadow, err = json.Marshal(shadow)
	if err != nil {
		return
	}
	if err = entity.GetDB().Save(d).Error; err != nil {
		return
	}

	// 如果是子设备，则对父设备的物模型进行一次调整
	isChildDevice, pIdentity, _ := utils.ParserIdentity(d.Identity)
	if isChildDevice {
		pDevice, _ := entity.GetPluginDevice(d.AreaID, d.PluginID, pIdentity)
		_ = UpdateShadowReported(pDevice, attr)
	}

	return
}

// GetControlAttributes 获取设备属性（不包括设备型号、厂商等属性）
func GetControlAttributes(d entity.Device) (attributes []entity.Attribute, err error) {
	das, err := GetThingModel(d)
	if err != nil {
		return
	}

	for _, instance := range das.Instances {
		if instance.Type == "info" {
			continue
		}
		// 如果是父设备，不需要显示子设备的instance出来
		if d.PID == 0 && instance.Attributes[0].Attribute.Attribute == "is_child_instance" {
			continue
		}

		as := plugin.GetInstanceControlAttributes(instance)
		attributes = append(attributes, as...)
	}
	return
}

// GetInstances 获取用户设备的属性(包括权限)
func GetInstances(areaID uint64, userID int, pluginID, identity string) (das plugin.DeviceInstances, err error) {
	d, err := entity.GetPluginDevice(areaID, pluginID, identity)
	if err != nil {
		return
	}

	return GetUserDeviceInstances(userID, d)
}

// GetUserDeviceInstances 获取用户设备的属性(包括权限)
func GetUserDeviceInstances(userID int, d entity.Device) (das plugin.DeviceInstances, err error) {

	if !plugin.GetGlobalClient().IsOnline(d) {
		err = errors2.New(status.DeviceOffline)
		return
	}

	up, err := entity.GetUserPermissions(userID)
	if err != nil {
		return
	}

	das, err = GetOrInitDeviceInstances(d, up)
	return
}

func GetOrInitDeviceInstances(d entity.Device, up entity.UserPermissions) (das plugin.DeviceInstances, err error) {
	if !d.IsInit() {
		if err = SyncDevice(&d); err != nil {
			return
		}
	}
	return GetDeviceInstances(d, up)
}

// GetDeviceInstances 获取设备的instance
func GetDeviceInstances(d entity.Device, up entity.UserPermissions) (das plugin.DeviceInstances, err error) {

	das, err = GetThingModel(d)
	if err != nil {
		return
	}

	for i, instance := range das.Instances {
		for j, a := range instance.Attributes {
			if up.IsDeviceAttrControlPermit(d.ID, instance.InstanceId, a.Attribute.Attribute) {
				das.Instances[i].Attributes[j].CanControl = true
			}
		}
	}

	// 如果是子设备，判断是否在线需要切换成父设备
	// 子设备的缓存属性在父设备中也存在，因此可以替换d
	if d.PID > 0 {
		d, _ = entity.GetDeviceByID(d.PID)
	}

	// 在线则直接使用设备影子中缓存的属性
	shadow, err := GetShadow(d)
	if err != nil {
		return
	}

	for i, ins := range das.Instances {
		for j, attr := range ins.Attributes {
			das.Instances[i].Attributes[j].Val, err = shadow.Get(ins.InstanceId, attr.Attribute.Attribute)
			if err != nil {
				return
			}
		}
	}
	return
}

// SyncDevice 同步设备属性
func SyncDevice(d *entity.Device) (err error) {
	var thingModel plugin.DeviceInstances
	thingModel, err = plugin.GetGlobalClient().GetAttributes(*d)
	if err != nil {
		return
	}
	// 保存物模型
	d.ThingModel, err = json.Marshal(thingModel)
	if err != nil {
		return
	}

	shadow := entity.NewShadow()
	for _, ins := range thingModel.Instances {
		for _, attr := range ins.Attributes {
			shadow.UpdateReported(ins.InstanceId, attr.Attribute)
		}
	}
	d.Shadow, err = json.Marshal(shadow)
	if err != nil {
		return
	}

	db := entity.GetDB()
	if err = db.Save(d).Error; err != nil {
		return
	}
	// 如果设备属性里面有子设备，则这里需要新增子设备
	if err = BatchInsertChildDevice(d, thingModel, entity.GetDB()); err != nil {
		return
	}

	return AddDevicePermissionForRoles(*d, db)
}

func UpdateDeviceShadow(em event.EventMessage) error {

	attr := em.GetAttr()
	if attr == nil {
		logger.Warn(" attr is nil")
		return nil
	}
	dID := em.GetDeviceID()
	d, err := entity.GetDeviceByID(dID)
	if err != nil {
		return err
	}
	// 从设备影子中获取属性
	shadow, err := GetShadow(d)
	if err != nil {
		return err
	}
	shadow.UpdateReported(attr.InstanceID, attr.Attribute)
	d.Shadow, err = json.Marshal(shadow)
	if err != nil {
		return err
	}
	if err = entity.GetDB().Save(d).Error; err != nil {
		return err
	}

	// 如果是子设备，则对父设备的物模型进行一次调整
	isChildDevice, pIdentity, _ := utils.ParserIdentity(d.Identity)
	if isChildDevice {
		pDevice, _ := entity.GetPluginDevice(d.AreaID, d.PluginID, pIdentity)
		em.SetDeviceID(pDevice.ID)
		_ = UpdateDeviceShadow(em)
	}
	return nil
}
