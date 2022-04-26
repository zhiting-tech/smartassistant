package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与场景相关的响应状态码
const (
	SceneNameExist = iota + 4000
	SceneNameSizeLimit
	SceneConditionNotExist
	SceneNotExist
	SceneCreateDeny
	SceneDeleteDeny
	SceneTypeForbidModify
	ConditionMisMatchTypeAndConfigErr
	ConditionTimingCountErr
	TaskTypeErr
	DeviceActionErr
	DeviceOperationNotSetErr
	DeviceOrSceneControlDeny
	DeviceOffline
	SceneParamIncorrectErr
	ConditionOfDeviceAttrWithoutReadPermission
	ConditionOfDeviceAttrWithoutNotifyPermission
)

func init() {
	errors.NewCode(SceneNameExist, "与其他场景重名,请修改")
	errors.NewCode(SceneNameSizeLimit, "场景名称长度不能超过40")
	errors.NewCode(SceneConditionNotExist, "场景触发条件不存在")
	errors.NewCode(SceneNotExist, "场景不存在")
	errors.NewCode(SceneCreateDeny, "您没有创建场景的权限")
	errors.NewCode(SceneDeleteDeny, "您没有删除场景的权限")
	errors.NewCode(SceneTypeForbidModify, "场景类型不允许修改")
	errors.NewCode(ConditionMisMatchTypeAndConfigErr, "场景触发条件类型与配置不一致")
	errors.NewCode(ConditionTimingCountErr, "定时触发条件只能添加一个")
	errors.NewCode(TaskTypeErr, "任务类型错误")
	errors.NewCode(DeviceActionErr, "设备操作类型不存在")
	errors.NewCode(DeviceOperationNotSetErr, "设备操作未设置")
	errors.NewCode(DeviceOrSceneControlDeny, "没有场景或设备的控制权限")
	errors.NewCode(DeviceOffline, "%s设备断连")
	errors.NewCode(SceneParamIncorrectErr, "%s不正确")
	errors.NewCode(ConditionOfDeviceAttrWithoutReadPermission, "场景触发条件的设备属性没有读权限")
	errors.NewCode(ConditionOfDeviceAttrWithoutNotifyPermission, "场景触发条件的设备属性没有通知权限")
}
