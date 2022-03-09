package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

// ChildInstances 子设备
type ChildInstances struct {
	Instances []interface{}
}

func (c ChildInstances) InstanceName() string {
	return "child_instances"
}

// ChildInstanceService 子设备服务
type ChildInstanceService interface {
	GetAid() int // 获取设备ID
	GetIid() int // 获取特征ID
}

// IsChildInstance 是否子设备
type IsChildInstance struct {
	attribute.Bool
}

func NewIsChildInstance() *IsChildInstance {
	return &IsChildInstance{}
}