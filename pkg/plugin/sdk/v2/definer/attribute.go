package definer

import (
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

func NewAttribute(attribute thingmodel.Attribute) *Attribute {
	return &Attribute{
		meta: &attribute,
	}
}

type Attribute struct {
	meta       *thingmodel.Attribute
	iAttribute thingmodel.IAttribute
}

func (a *Attribute) SetVal(val interface{}) {
	a.meta.Val = val
}

func (a *Attribute) GetVal() (val interface{}) {
	return a.meta.Val
}

func (a *Attribute) Set(val interface{}) error {
	if a.iAttribute == nil {
		return NotEnableErr
	}
	return a.iAttribute.Set(val)
}

// Enable 启用属性并通过实现接口设置方法
func (a *Attribute) Enable(attr thingmodel.IAttribute) *Attribute {
	a.iAttribute = attr
	return a
}

// SetRange 设置最小值最大值
func (a *Attribute) SetRange(min, max interface{}) *Attribute {
	a.meta.Min = min
	a.meta.Max = max
	return a
}

// Type 属性类型
func (a *Attribute) Type() thingmodel.Attribute {
	return *a.meta
}
