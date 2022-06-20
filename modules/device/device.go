package device

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

func Create(areaID uint64, device *entity.Device) (err error) {
	if device == nil {
		return errors.New("nil device")
	}

	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {

		device.AreaID = areaID
		if device.IsSa() {
			// 添加设备为SA时不需要添加设备影子
			if err = entity.CreateSA(device, tx); err != nil {
				return err
			}
		} else {
			if err = entity.CreateDevice(device, tx); err != nil {
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
		if entity.IsDeviceActionPermit(role.ID, types.ActionControl, tx) {
			role.AddPermissionsWithDB(tx, ManagePermissions(device)...)
		}

		if entity.IsDeviceActionPermit(role.ID, types.ActionUpdate, tx) {
			role.AddPermissionsWithDB(tx, types.NewDeviceUpdate(device.ID))
		}

		// SA设备不需要配置控制和删除权限
		if device.IsSa() {
			continue
		}
		if entity.IsDeviceActionPermit(role.ID, types.ActionControl, tx) {
			var ps []types.Permission
			ps, err = ControlPermissions(device, true)
			if err != nil {
				logger.Error("ControlPermissionsErr:", err.Error())
				continue
			}
			role.AddPermissionsWithDB(tx, ps...)
		}

		if entity.IsDeviceActionPermit(role.ID, types.ActionDelete, tx) {
			role.AddPermissionsWithDB(tx, types.NewDeviceDelete(device.ID))
		}
	}
	return
}

// GetThingModel 获取用户设备的属性(包括权限)
func GetThingModel(areaID uint64, pluginID, iid string) (das thingmodel.ThingModel, err error) {
	d, err := entity.GetPluginDevice(areaID, pluginID, iid)
	if err != nil {
		return
	}

	return d.GetThingModel()
}

// LogoURL 设备Logo图片地址
func LogoURL(req *http.Request, d entity.Device) string {
	return plugin.DeviceLogoURL(req, d.PluginID, d.Model, d.Type)
}

// LogoInfo 获取logo的信息
func LogoInfo(c *gin.Context, d entity.Device) (logoUrl, logo string) {
	if d.LogoType == nil || *d.LogoType == int(types.NormalLogo) {
		return LogoURL(c.Request, d), plugin.GetGlobalClient().Config(d.PluginID).DeviceConfig(d.Model, d.Type).Logo
	}
	return types.LogoFromLogoType(types.LogoType(*d.LogoType), c.Request)

}

func TypeToLogoType(deviceType plugin.DeviceType) types.LogoType {

	switch deviceType {
	case plugin.TypeLight:
		return types.LightLogo
	case plugin.TypeSwitch:
		return types.SwitchLogo
	case plugin.TypeOutlet:
		return types.OutletLogo
	case plugin.TypeCurtain:
		return types.CurtainLogo
	case plugin.TypeGateway:
		return types.GatewayLogo
	case plugin.TypeSensor:
		return types.SensorLogo
	case plugin.TypeCamera:
		return types.CameraLogo
	case plugin.TypeDoorLock:
		return types.DoorLockLogo
	case plugin.TypeTemperatureAndHumiditySensor:
		return types.TemperatureAndHumiditySensorLogo
	case plugin.TypeWaterLeakSensor:
		return types.WaterLeakSensorLogo
	case plugin.TypeHumanSensors:
		return types.HumanSensorLogo
	case plugin.TypeWindowDoorSensor:
		return types.WindowDoorSensorLogo
	case plugin.TypeSmokeSensor:
		return types.SmokeSensorLogo
	case plugin.TypeGasSensor:
		return types.GasSensorLogo
	default:
		return types.OthLogo
	}
}
