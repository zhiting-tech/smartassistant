package server

import (
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/utils"
)

type DeviceInfo struct {
	Identity     string
	Model        string
	Manufacturer string
	AuthRequired bool
}

type Notification struct {
	Identity   string
	InstanceID int
	Attr       string
	Val        interface{}
}

type WatchChan chan Notification

type Device interface {
	Identity() string
	Info() DeviceInfo
	Setup() error
	Online() bool
	Update() error
	Close() error
	GetChannel() WatchChan
}

// AuthDevice 需要认证的设备
type AuthDevice interface {
	Device
	// IsAuth 返回设备是否成功认证/配对
	IsAuth() bool
	// Auth 认证/配对
	Auth(params map[string]string) error
	// RemoveAuthorization 取消认证/配对
	RemoveAuthorization(params map[string]string) error
}

// ParentDevice 父设备：拥有子设备的设备
type ParentDevice interface {
	Device
	GetChildDeviceById(instanceId int) ChildDevice
}

// ChildDevice 子设备，暂时只需要Online方法
type ChildDevice interface {
	Online() bool
}

type OTAProgressState int // OTA进度

const (
	OTAFinish OTAProgressState = 100 // OTA成功

	OTAUpgradeFail  OTAProgressState = -1 // 更新失败
	OTADownloadFail OTAProgressState = -2 // 下载失败
	OTAValidateFail OTAProgressState = -3 // 校验失败
	OTABurnFail     OTAProgressState = -4 // 烧写失败
)

func OTAProgress(i int) OTAProgressState {
	return OTAProgressState(i)
}

type OTAResp struct {
	Step OTAProgressState
}

type OTADevice interface {
	Device
	OTA(firmwareURL string) (chan OTAResp, error)
}

type Notify struct {
	Identity   string    `json:"identity"`
	InstanceID int       `json:"instance_id"`
	Attribute  Attribute `json:"attribute"`
}

type Attribute struct {
	ID         int         `json:"id"`
	Attribute  string      `json:"attribute"`
	Val        interface{} `json:"val"`
	ValType    string      `json:"val_type"`
	Min        *int        `json:"min,omitempty"`
	Max        *int        `json:"max,omitempty"`
	Permission uint        `json:"permission"`
}

type Instance struct {
	Type       string      `json:"type"`
	InstanceId int         `json:"instance_id"`
	Attributes []Attribute `json:"attributes"`
}

func GetDeviceInstances(device Device) (instances []Instance) {

	// parse device
	d := utils.Parse(device)
	logrus.Debugf("total %d instances\n", len(d.Instances))
	for _, ins := range d.Instances {

		var attrs []Attribute
		logrus.Debugf("total %d attrs of instance %d\n", len(ins.Attributes), ins.ID)
		for _, attr := range ins.Attributes {
			if attr == nil || !attr.Require && !attr.Active {
				logrus.Debug("attr is nil or not active")
				continue
			}
			a := Attribute{
				ID:         attr.ID,
				Attribute:  attr.Name,
				Val:        attribute.ValueOf(attr.Model),
				ValType:    attr.Type,
				Permission: attr.Permission,
			}
			if num, ok := attr.Model.(attribute.IntType); ok {
				a.Min, a.Max = num.GetRange()
			}

			attrs = append(attrs, a)
		}

		instance := Instance{
			Type:       ins.Type,
			InstanceId: ins.ID,
			Attributes: attrs,
		}
		instances = append(instances, instance)
	}
	return
}
