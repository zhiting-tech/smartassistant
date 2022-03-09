package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

// StatelessSwitch 无状态开关
type StatelessSwitch struct {
	IsChildInstance *IsChildInstance
	Name            *attribute.Name
	Model        	*attribute.Model
	Manufacturer 	*attribute.Manufacturer
	Version      	*attribute.Version

	SwitchEvent		*SwitchEvent
	Battery 		*Battery
}

func (s StatelessSwitch) InstanceName() string {
	return "wireless_switch"
}

// SwitchEvent 开关事件
type SwitchEvent struct {
	attribute.Int
}

func NewSwitchEvent() *SwitchEvent {
	return &SwitchEvent{}
}