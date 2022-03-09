package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

type Outlet struct {
	Power *attribute.Power `tag:"required"`
	InUse *InUse
}

func (o Outlet) InstanceName() string {
	return "outlet"
}

// InUse 电源插座是否有物理插入
type InUse struct {
	attribute.Bool
}

func NewInUse() *InUse {
	return &InUse{}
}
