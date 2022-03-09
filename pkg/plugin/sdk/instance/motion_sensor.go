package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

// MotionSensor 人体传感器
type MotionSensor struct {
	IsChildInstance *IsChildInstance
	Name            *attribute.Name
	Model        	*attribute.Model
	Manufacturer 	*attribute.Manufacturer
	Version      	*attribute.Version

	Detected *Detected
	Battery *Battery
}

func (w MotionSensor) InstanceName() string {
	return "human_sensors"
}

type Detected struct {
	attribute.Int
}

// NewDetected 侦查
func NewDetected() *Detected {
	return &Detected{}
}

type Battery struct {
	attribute.Int
}

// NewBattery 电量
func NewBattery() *Battery {
	return &Battery{}
}
