package device

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// IsPermit 判断用户是否有权限
func IsPermit(c *gin.Context, p types.Permission) bool {
	u := session.Get(c)
	return u != nil && entity.JudgePermit(u.UserID, p)
}

// ControlPermissions 根据配置获取设备所有控制权限
func ControlPermissions(d entity.Device, withHidden bool) ([]types.Permission, error) {
	as, err := d.ControlAttributes(withHidden)
	if err != nil {
		logger.Error("GetControlAttributesErr", err)
		return nil, err
	}

	target := types.DeviceTarget(d.ID)
	res := make([]types.Permission, 0)
	for _, attr := range as {
		name := attr.Attribute.Type
		p := types.Permission{
			Name:      name,
			Action:    types.ActionControl,
			Target:    target,
			Attribute: strconv.Itoa(attr.AID),
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

	controlPermission, err := ControlPermissions(d, false)
	if err != nil {
		return
	}

	// 非SA设备可配置删除设备权限,控制设备权限
	ps = append(ps, controlPermission...)
	ps = append(ps, types.NewDeviceDelete(d.ID))
	return
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
