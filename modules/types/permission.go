package types

import (
	"fmt"
)

const (
	FwUpgrade       = "firmware_upgrade" // 固件升级
	SoftwareUpgrade = "software_upgrade" // 软件升级
)

const (
	ActionAdd     = "add"
	ActionGet     = "get"
	ActionUpdate  = "update"
	ActionControl = "control"
	ActionDelete  = "delete"
	ActionManage  = "manage"
)

type Permission struct {
	Name      string `json:"name"`
	Action    string `json:"action"`    // 动作
	Target    string `json:"target"`    // 对象
	Attribute string `json:"attribute"` // 属性
}

// 设备
var (
	DeviceAdd         = Permission{"添加设备", ActionAdd, "device", ""}
	DeviceUpdate      = Permission{"修改设备", ActionUpdate, "device", ""}
	DeviceControl     = Permission{"控制设备", ActionControl, "device", ""}
	DeviceDelete      = Permission{"删除设备", ActionDelete, "device", ""}
	DeviceUpdateOrder = Permission{"设备排序", ActionUpdate, "device", "order"}
)

// 家庭/公司
var (
	AreaGetCode          = Permission{"生成邀请码", ActionGet, "area", "invite_code"}
	AreaUpdateName       = Permission{"修改家庭名称", ActionUpdate, "area", "name"}
	AreaUpdateMemberRole = Permission{"修改成员角色", ActionUpdate, "area", "member_role"}
	AreaDelMember        = Permission{"删除成员", ActionDelete, "area", "member"}
)

// 公司
var (
	AreaUpdateMemberDepartment = Permission{"修改成员部门", ActionUpdate, "area", "member_department"}
	AreaUpdateCompanyName      = Permission{"修改公司名称", ActionUpdate, "area", "company_name"}
)

// 房间/区域
var (
	LocationAdd         = Permission{"添加房间/区域", ActionAdd, "location", ""}
	LocationUpdateOrder = Permission{"调整顺序", ActionUpdate, "location", "order"}
	LocationUpdateName  = Permission{"修改房间名称", ActionUpdate, "location", "name"}
	LocationGet         = Permission{"查看房间详情", ActionGet, "location", ""}
	LocationDel         = Permission{"删除房间", ActionDelete, "location", ""}
)

// 角色
var (
	RoleGet    = Permission{"查看角色列表", ActionGet, "role", ""}
	RoleAdd    = Permission{"新增角色", ActionAdd, "role", ""}
	RoleUpdate = Permission{"编辑角色", ActionUpdate, "role", ""}
	RoleDel    = Permission{"删除角色", ActionDelete, "role", ""}
)

// 场景
var (
	SceneAdd     = Permission{"新增场景", ActionAdd, "scene", ""}
	SceneUpdate  = Permission{"修改场景", ActionUpdate, "scene", ""}
	SceneDel     = Permission{"删除场景", ActionDelete, "scene", ""}
	SceneControl = Permission{"控制场景", ActionControl, "scene", ""}
)

// 部门
var (
	DepartmentAdd         = Permission{"添加部门", ActionAdd, "department", ""}
	DepartmentUpdateOrder = Permission{"调整部门顺序", ActionUpdate, "department", "order"}
	DepartmentGet         = Permission{"查看部门详情", ActionGet, "department", ""}
	DepartmentAddUser     = Permission{"添加成员", ActionAdd, "department", "user"}
	DepartmentUpdate      = Permission{"部门设置", ActionUpdate, "department", ""}
)

var (
	DevicePermission     = []Permission{DeviceAdd, DeviceUpdate, DeviceControl, DeviceDelete, DeviceUpdateOrder}
	AreaPermission       = []Permission{AreaGetCode, AreaUpdateName, AreaUpdateMemberRole, AreaDelMember}
	LocationPermission   = []Permission{LocationAdd, LocationUpdateOrder, LocationUpdateName, LocationGet, LocationDel}
	RolePermission       = []Permission{RoleGet, RoleAdd, RoleUpdate, RoleDel}
	ScenePermission      = []Permission{SceneAdd, SceneUpdate, SceneDel, SceneControl}
	DepartmentPermission = []Permission{DepartmentAdd, DepartmentUpdateOrder, DepartmentGet, DepartmentAddUser, DepartmentUpdate}
	CompanyPermission    = []Permission{AreaGetCode, AreaUpdateCompanyName, AreaUpdateMemberRole, AreaUpdateMemberDepartment, AreaDelMember}
)

var (
	DefaultPermission []Permission
	// ManagerPermission 管理员默认权限
	ManagerPermission []Permission
	// MemberPermission 成员默认权限
	MemberPermission []Permission
)

func DeviceTarget(deviceID int) string {
	return fmt.Sprintf("device-%d", deviceID)
}

func NewDeviceDelete(deviceID int) Permission {
	target := DeviceTarget(deviceID)
	return Permission{"删除设备", ActionDelete, target, ""}
}
func NewDeviceUpdate(deviceID int) Permission {
	target := DeviceTarget(deviceID)
	return Permission{"修改设备", ActionUpdate, target, ""}
}

func NewDeviceManage(deviceID int, name string, attr string) Permission {
	target := DeviceTarget(deviceID)
	return Permission{name, ActionManage, target, attr}
}

func NewDeviceFwUpgrade(deviceID int) Permission {
	return NewDeviceManage(deviceID, "固件升级", FwUpgrade)
}

func NewDeviceSoftwareUpgrade(deviceID int) Permission {
	return NewDeviceManage(deviceID, "软件升级", SoftwareUpgrade)
}

func init() {

	DefaultPermission = append(DefaultPermission, DevicePermission...)
	DefaultPermission = append(DefaultPermission, AreaPermission...)
	DefaultPermission = append(DefaultPermission, LocationPermission...)
	DefaultPermission = append(DefaultPermission, RolePermission...)
	DefaultPermission = append(DefaultPermission, ScenePermission...)
	DefaultPermission = append(DefaultPermission, DepartmentPermission...)
	DefaultPermission = append(DefaultPermission, AreaUpdateMemberDepartment, AreaUpdateCompanyName)

	ManagerPermission = append(ManagerPermission, DefaultPermission...)
	MemberPermission = []Permission{DeviceControl, LocationGet, DepartmentGet}
}
