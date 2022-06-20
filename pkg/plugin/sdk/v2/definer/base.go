package definer

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

func NewService(serviceType thingmodel.ServiceType) *BaseService {

	b := BaseService{
		serviceType:  serviceType,
		attributeMap: make(map[string]*Attribute),
	}

	return &b
}

type BaseService struct {
	serviceType  thingmodel.ServiceType
	notifyFunc   NotifyFunc
	iid          string
	_instance    *Instance
	attributeMap map[string]*Attribute
}

func (b BaseService) Type() thingmodel.ServiceType {
	return b.serviceType
}

func (b *BaseService) WithAttributes(attrTypes ...thingmodel.Attribute) *BaseService {

	for _, attrType := range attrTypes {
		b.WithAttribute(attrType)
	}
	return b
}

func (b *BaseService) WithAttribute(attr thingmodel.Attribute) *Attribute {

	a := NewAttribute(attr)
	b.attributeMap[attr.String()] = a
	b._instance.AddAttribute(a)
	if attr.Default != nil {
		a.meta.Val = attr.Default
	} else {
		switch a.meta.ValType {
		case thingmodel.String:
			a.meta.Val = ""
		case thingmodel.Int:
			a.meta.Val = 0
		case thingmodel.Int32:
			a.meta.Val = 0
		case thingmodel.Int64:
			a.meta.Val = 0
		case thingmodel.Bool:
			a.meta.Val = false
		case thingmodel.Float32:
			a.meta.Val = 0
		case thingmodel.Float64:
			a.meta.Val = 0
		case thingmodel.Enum:
			a.meta.Val = 0
		case thingmodel.JSON:
			a.meta.Val = ""
		}
	}
	return a
}

func (b *BaseService) Enable(attrType thingmodel.Attribute, getSetter thingmodel.IAttribute) *Attribute {
	a := b.attributeMap[attrType.String()]
	if a == nil {
		a = b.WithAttribute(attrType)
	}
	a.Enable(getSetter)
	return a
}

func (b BaseService) Notify(attrType thingmodel.Attribute, val interface{}) error {
	a := b.attributeMap[attrType.String()]
	if a == nil {
		return fmt.Errorf("notify err, attr:%s not found of %s:%s",
			attrType.String(), b.iid, b.Type())
	}
	a.SetVal(val)
	if b.notifyFunc != nil {
		aid := a.meta.AID
		ae := AttributeEvent{
			IID: b.iid,
			AID: aid,
			Val: val,
		}
		return b.notifyFunc(ae)
	}
	logrus.Warnf("attribute notify function not set: %s", attrType.String())
	return nil
}

func (b *BaseService) SetNotifyFunc(iid string, nf NotifyFunc) {
	b.notifyFunc = nf
	b.iid = iid
}

func (b *BaseService) GetAttribute(attrType thingmodel.Attribute) *Attribute {
	return b.attributeMap[attrType.String()]
}
