package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与设备相关的响应状态码
const (
	DeviceExist = iota + 2000
	SaDeviceAlreadyBind
	DeviceNameInputNilErr
	DeviceNameLengthLimit
	DeviceNotExist
	NotBoundDevice
	DataSyncFail
	AlreadyDataSync
	ForbiddenBindOtherSA
	ForbiddenRemoveSADevice
	DeviceTypeNotExist
)

func init() {
	errors.NewCode(DeviceExist, "设备已被添加")
	errors.NewCode(SaDeviceAlreadyBind, "设备已被绑定")
	errors.NewCode(DeviceNameInputNilErr, "请输入设备名称")
	errors.NewCode(DeviceNameLengthLimit, "设备名称长度不能超过20")
	errors.NewCode(DeviceNotExist, "该设备不存在")
	errors.NewCode(NotBoundDevice, "当前用户未绑定该设备")
	errors.NewCode(DataSyncFail, "数据同步失败,请重试")
	errors.NewCode(AlreadyDataSync, "数据已同步,禁止多次同步数据")
	errors.NewCode(ForbiddenBindOtherSA, "已有SA，不允许添加其他SA")
	errors.NewCode(ForbiddenRemoveSADevice, "不允许删除SA设备")
	errors.NewCode(DeviceTypeNotExist, "设备类型不存在")
}
