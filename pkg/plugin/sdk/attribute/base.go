package attribute

import (
	"github.com/sirupsen/logrus"
)

const (
	AttrPermissionALLControl = iota // 0：可控制可读(例如灯光控制属性)
	AttrPermissionOnlyRead          // 1：不可控制但是可读(例如：传感器状态属性)
	AttrPermissionNone              // 2：不可控制不可读(其他属性)
)

type UpdateFunc func(val interface{}) error
type NotifyFunc func(val interface{}) error

type Setter interface {
	Set(val interface{}) error
}

type Notifier interface {
	Notify(val interface{}) error
	SetNotifyFunc(NotifyFunc)
}

// HomekitCharacteristic homekit的特征属性
type HomekitCharacteristic interface {
	SetAid(aid uint64)
	GetAid() uint64
	SetIid(iid uint64)
	GetIid() uint64
}

// AttrPermission 属性状态接口
type AttrPermission interface {
	SetPermission(permission uint)
	GetPermission() uint
}

type Base struct {
	updateFn   UpdateFunc
	notifyFunc NotifyFunc
	Permission uint   // 属性状态： 0可控制可读(灯光控制属性)，1不可控制但是可读(传感器状态)，2不可控制不可读(其他属性)
	aid        uint64 // homekit: Accessory Instance ID 设备ID
	iid        uint64 // homekit: Instance ID 特征ID
}

// SetUpdateFunc 设置属性更新函数
func (b *Base) SetUpdateFunc(fn UpdateFunc) {
	b.updateFn = fn
}

// Set 触发Base.updateFn，更新设备属性
func (b *Base) Set(val interface{}) error {
	defer func() {
		if err := recover(); err != nil {
			logrus.Error("set err:", err)
		}
	}()

	if b.updateFn != nil {
		return b.updateFn(val)
	}

	logrus.Warn("update func not set")

	return nil
}

// SetNotifyFunc  设置通知函数
func (b *Base) SetNotifyFunc(fn NotifyFunc) {
	if b.notifyFunc == nil {
		b.notifyFunc = fn
	}
}

// Notify 触发Base.notifyFn,通过channel通知SA
func (b *Base) Notify(val interface{}) error {
	if b.notifyFunc != nil {
		return b.notifyFunc(val)
	}

	logrus.Warn("notify func not set")

	return nil
}

func (b *Base) SetAid(aid uint64) {
	b.aid = aid
}

func (b *Base) GetAid() uint64 {
	return b.aid
}

func (b *Base) SetIid(iid uint64) {
	b.iid = iid
}

func (b *Base) GetIid() uint64 {
	return b.iid
}

func (b *Base) SetPermission(permission uint) {
	b.Permission = permission
}

func (b *Base) GetPermission() uint {
	return b.Permission
}

func StringWithValidValues(values ...string) String {
	s := String{}
	if len(s.validValues) == 0 {
		s.validValues = make(map[string]interface{})
	}
	for _, values := range values {
		s.validValues[values] = struct{}{}
	}
	return s
}

func TypeOf(iface interface{}) string {
	switch iface.(type) {
	case IntType, EnumType:
		return "int"
	case BoolType:
		return "bool"
	case StringType:
		return "string"
	}
	return ""
}

func ValueOf(iface interface{}) interface{} {
	switch v := iface.(type) {
	case IntType:
		return v.GetInt()
	case BoolType:
		return v.GetBool()
	case StringType:
		return v.GetString()
	case EnumType:
		return v.GetEnum()
	}
	return ""
}
