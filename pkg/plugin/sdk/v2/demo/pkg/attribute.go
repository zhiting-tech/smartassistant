package pkg

import (
	"github.com/sirupsen/logrus"
	"net"
)

// OnOff 开关属性，为每个属性定义结构实现 Get(), Set()方法
type OnOff struct {
	conn net.Conn
}

func (s OnOff) Set(value interface{}) error {
	logrus.Printf("switch set %v", value)
	return nil
}

func (s OnOff) Get() (interface{}, error) {
	logrus.Println("switch get")
	return "off", nil
}

type Brightness struct {
}

func (s Brightness) Set(value interface{}) error {
	logrus.Printf("brightness set %v", value)
	return nil
}

func (s Brightness) Get() (interface{}, error) {
	logrus.Println("brightness get")
	return 97, nil
}

// NewAttribute 属性多的情况下可以使用一个struct实现，避免过多定义struct
func NewAttribute(attributeType string, alias string) Attribute {
	return Attribute{_type: attributeType, alias: alias}
}

type Attribute struct {
	_type string
	alias string
}

func (s Attribute) Set(value interface{}) error {
	logrus.Printf("switch set %s %v", s._type, value)
	return nil
}

func (s Attribute) Get() (interface{}, error) {
	logrus.Printf("switch get %s", s._type)
	return "off", nil
}
