package utils

import (
	"fmt"
	"go/ast"
	"reflect"
	"regexp"
	"strings"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/instance"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// Attribute represents a column of database
type Attribute struct {
	ID         int
	Model      interface{}
	TypeModel  interface{}
	Name       string
	Type       string
	Tag        string
	Require    bool
	Active     bool
	Aid        uint64 // homekit aid
	Iid        uint64 // homekit iid
	Permission uint   // 控制状态： 0可控制可读(灯光控制属性)，1不可控制但是可读(传感器状态)，2不可控制不可读(其他属性)
}

func (a *Attribute) parseTag() {

	strs := strings.Split(a.Tag, ";")
	for _, v := range strs {
		infoStr := strings.Split(v, ":")
		if len(infoStr) != 0 {
			switch infoStr[0] {
			case "name":
			case "required":
				a.Require = true
			}
		}
	}
	return
}

type Instance struct {
	ID             int
	Model          interface{}
	Name           string
	Type           string
	Tag            string
	Attributes     []*Attribute
	AttributeNames []string
	AttributeMap   map[string]*Attribute
}

// GetAttribute return instance by name
func (instance *Instance) GetAttribute(name string) *Attribute {
	return instance.AttributeMap[name]
}

// Device represents a table of database
type Device struct {
	Model         interface{}
	Instances     []*Instance
	InstanceNames []string
	InstanceMap   map[int]*Instance
}

// GetAttribute return Attribute by id
func (d *Device) GetAttribute(instanceID int, attr string) *Attribute {
	if ins, ok := d.InstanceMap[instanceID]; ok {
		if attr, ok := ins.AttributeMap[attr]; ok {
			return attr
		}
	}
	return nil
}

type DeviceName interface {
	DeviceName() string
}
type InstanceName interface {
	InstanceName() string
}

// Parse a struct to a Device instance
func Parse(dest interface{}) *Device {
	if dest == nil {
		return nil
	}
	deviceType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	deviceValue := reflect.Indirect(reflect.ValueOf(dest))
	device := &Device{
		Model:       dest,
		InstanceMap: make(map[int]*Instance),
	}
	instanceNum := 1
	for i := 0; i < deviceType.NumField(); i++ {
		p := deviceType.Field(i)
		v := reflect.Indirect(deviceValue.Field(i))
		if !p.Anonymous && ast.IsExported(p.Name) {
			t, ok := v.Interface().(InstanceName)
			if !ok {
				logrus.Warnln(p.Name, "is not a instance")
				continue
			}
			destArr := []interface{}{v.Interface()}

			// 如果是子设备的物模型，需要把子设备的数据解析出来
			if t.InstanceName() == "child_instances" {
				destArr = v.Interface().(instance.ChildInstances).Instances
			}
			// 如果是网关设备的物模型，需要把网关设备的数据解析出来
			if t.InstanceName() == "gateway_services" {
				destArr = v.Interface().(instance.GatewayServices).Services
			}

			for _, dest := range destArr {
				ins := ParseInstance(dest)
				if v, ok := p.Tag.Lookup("tag"); ok {
					ins.Tag = v
				}
				ins.ID = instanceNum
				device.Instances = append(device.Instances, ins)
				device.InstanceNames = append(device.InstanceNames, p.Name)
				device.InstanceMap[ins.ID] = ins
				instanceNum++
			}
		}
	}
	return device
}

// ParseInstance a instance to a Instance instance
func ParseInstance(dest interface{}) *Instance {
	instanceType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	var instanceName string
	t, ok := dest.(InstanceName)
	if !ok {
		instanceName = instanceType.Name()
	} else {
		instanceName = t.InstanceName()
	}

	ins := &Instance{
		Model:        dest,
		Name:         ToSnakeCase(instanceName),
		AttributeMap: make(map[string]*Attribute),
		Type:         ToSnakeCase(instanceName),
	}
	attrs := DeepAttrs(dest, 1)

	ins.Attributes = append(ins.Attributes, attrs...)
	for _, attr := range attrs {
		ins.AttributeNames = append(ins.AttributeNames, attr.Name)
		ins.AttributeMap[ToSnakeCase(attr.Name)] = attr
	}
	return ins
}

// DeepAttrs 递归获取所有属性
func DeepAttrs(dest interface{}, incrAttrID int) (attrs []*Attribute) {

	instanceType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	instanceValue := reflect.Indirect(reflect.ValueOf(dest))

	for i := 0; i < instanceType.NumField(); i++ {
		t := instanceType.Field(i)
		v := instanceValue.Field(i)
		if t.Anonymous { // 匿名结构体则递归获取其属性
			deepAttrs := DeepAttrs(v.Interface(), incrAttrID)
			attrs = append(attrs, deepAttrs...)
			incrAttrID += len(deepAttrs)
		} else if v.Kind() == reflect.Slice { // 切片则递归获取其属性
			sliceLen := v.Len()
			// 循环切片里的每一个服务
			for k := 0; k < sliceLen; k++ {
				value := v.Index(k)
				attr := Attribute{
					ID:   incrAttrID,
					Name: fmt.Sprintf("%s_%d", ToSnakeCase(t.Name), k+1),
					Type: attribute.TypeOf(reflect.Indirect(reflect.New(value.Type())).Interface()), // FIXME
				}
				setAttrByValue(&attr, value)
				attrs = append(attrs, &attr)
				incrAttrID++
			}
		} else if ast.IsExported(t.Name) { // 导出字段作为属性
			attr := Attribute{
				ID:   incrAttrID,
				Name: ToSnakeCase(t.Name),
				Type: attribute.TypeOf(reflect.Indirect(reflect.New(t.Type)).Interface()), // FIXME
			}
			if v.Kind() != reflect.Ptr {
				logrus.Warnln(instanceType.Name(), t.Name, "is not pointer, ignore")
				continue
			}
			if v.Kind() != reflect.Struct && !v.IsNil() {
				setAttrByValue(&attr, v)
			}
			if v, ok := t.Tag.Lookup("tag"); ok {
				attr.Tag = v
			}
			attr.parseTag()
			incrAttrID++
			attrs = append(attrs, &attr)
		}
	}
	return attrs
}

// setAttrByValue 通过reflect.Value设置attr的字段值
func setAttrByValue(attr *Attribute, value reflect.Value) {
	attr.Active = true
	attr.Model = value.Interface()
	// 赋值homekit所需要的aid跟iid
	homekitCharacteristic, ok := (value.Interface()).(attribute.HomekitCharacteristic)
	if ok {
		attr.Aid = homekitCharacteristic.GetAid()
		attr.Iid = homekitCharacteristic.GetIid()
	}
	// 赋值属性的状态
	attrPermission, ok := (value.Interface()).(attribute.AttrPermission)
	if ok {
		attr.Permission = attrPermission.GetPermission()
	}

	return
}
