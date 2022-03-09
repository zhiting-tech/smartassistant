package entity

import (
	"fmt"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/types"
)

type ActionType string

// RolePermission 角色权限role_tmpl
type RolePermission struct {
	ID        int
	RoleID    int  `gorm:"index:permission,unique"` // 角色
	Role      Role `gorm:"constraint:OnDelete:CASCADE;"`
	Name      string
	Action    string `gorm:"index:permission,unique"` // 动作
	Target    string `gorm:"index:permission,unique"` // 对象
	Attribute string `gorm:"index:permission,unique"` // 属性
}

func (p RolePermission) TableName() string {
	return "role_permissions"
}

// IsDeviceControlPermit 判断用户是否有该设备的某个控制权限
func IsDeviceControlPermit(userID, deviceID int, attr Attribute) bool {
	return IsDeviceControlPermitByAttr(userID, deviceID, attr.InstanceID, attr.Attribute.Attribute)
}

func PluginDeviceAttr(instanceID int, attr string) string {
	return fmt.Sprintf("%d_%s", instanceID, attr)
}

// IsDeviceControlPermitByAttr 判断用户是否有该设备的某个控制权限
func IsDeviceControlPermitByAttr(userID, deviceID, instanceID int, attr string) bool {
	target := types.DeviceTarget(deviceID)
	return judgePermit(userID, "control", target, PluginDeviceAttr(instanceID, attr))
}

type Attr struct {
	DeviceID   int
	InstanceID int
	Attribute  string
}

type UserPermissions struct {
	ps      []RolePermission
	isOwner bool
}

func (up UserPermissions) IsOwner() bool {
	return up.isOwner
}

// IsDeviceControlPermit 判断设备是否可控制
func (up UserPermissions) IsDeviceControlPermit(deviceID int) bool {
	if up.isOwner {
		return true
	}
	for _, p := range up.ps {
		if p.Action == "control" &&
			p.Target == types.DeviceTarget(deviceID) {
			return true
		}
	}
	return false
}

// IsDeviceAttrControlPermit 判断设备的属性是否有权限
func (up UserPermissions) IsDeviceAttrControlPermit(deviceID, instanceID int, attr string) bool {
	if up.isOwner {
		return true
	}
	for _, p := range up.ps {
		if p.Action == "control" &&
			p.Target == types.DeviceTarget(deviceID) &&
			p.Attribute == PluginDeviceAttr(instanceID, attr) {
			return true
		}
	}
	return false
}

// IsDeviceAttrPermit 判断设备的属性是否有权限
func (up UserPermissions) IsDeviceAttrPermit(deviceID int, attr Attribute) bool {
	if up.isOwner {
		return true
	}
	// 如果设备的属性是只读，例如传感器，则默认有权限（这种属性，不需要权限控制）
	if attr.Permission == attribute.AttrPermissionOnlyRead {
		return true
	}
	for _, p := range up.ps {
		if p.Action == "control" &&
			p.Target == types.DeviceTarget(deviceID) &&
			p.Attribute == PluginDeviceAttr(attr.InstanceID, attr.Attribute.Attribute) {
			return true
		}
	}

	return false
}

func (up UserPermissions) IsPermit(tp types.Permission) bool {
	if up.isOwner {
		return true
	}
	for _, p := range up.ps {
		if p.Action == tp.Action && p.Target == tp.Target && p.Attribute == tp.Attribute {
			return true
		}
	}
	return false
}

// GetUserPermissions 获取用户的所有权限
func GetUserPermissions(userID int) (up UserPermissions, err error) {
	var ps []RolePermission
	if err = GetDB().Scopes(UserRolePermissionsScope(userID)).
		Find(&ps).Error; err != nil {
		return
	}
	return UserPermissions{ps: ps, isOwner: IsOwner(userID)}, nil
}

func UserRolePermissionsScope(userID int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select("role_permissions.*").
			Joins("inner join roles on roles.id=role_permissions.role_id").
			Joins("inner join user_roles on user_roles.role_id=roles.id").
			Where("user_roles.user_id=?", userID)
	}
}
func JudgePermit(userID int, p types.Permission) bool {
	return judgePermit(userID, p.Action, p.Target, p.Attribute)
}

func judgePermit(userID int, action, target, attribute string) bool {
	// SA拥有者默认拥有所有权限
	if IsOwner(userID) {
		return true
	}

	var permissions []RolePermission
	if err := GetDB().Scopes(UserRolePermissionsScope(userID)).
		Where("action = ? and target = ? and attribute = ?",
			action, target, attribute).Find(&permissions).Error; err != nil {
		return false
	}

	if len(permissions) == 0 {
		return false
	}

	return true
}

func IsPermit(roleID int, action, target, attribute string, tx *gorm.DB) bool {
	p := RolePermission{
		RoleID:    roleID,
		Action:    action,
		Target:    target,
		Attribute: attribute,
	}
	if err := tx.First(&p, p).Error; err != nil {
		return false
	}
	return true
}

func IsDeviceActionPermit(roleID int, action string, tx *gorm.DB) bool {
	return IsPermit(roleID, action, "device", "", tx)
}
