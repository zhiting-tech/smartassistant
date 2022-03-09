package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

// WaterLeakSensor 水浸传感器
type WaterLeakSensor struct {
	IsChildInstance *IsChildInstance
	Name            *attribute.Name
	Model        	*attribute.Model
	Manufacturer 	*attribute.Manufacturer
	Version      	*attribute.Version

	LeakDetected *LeakDetected
	Battery *Battery
}

func (w WaterLeakSensor) InstanceName() string {
	return "water_leak_sensor"
}

// LeakDetected 0:表示未检测到水浸 1:表示检测到水浸
type LeakDetected struct {
	attribute.Int
}

func NewLeakDetected() *LeakDetected {
	return &LeakDetected{}
}