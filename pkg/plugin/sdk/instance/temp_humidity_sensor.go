package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

// TempHumiditySensor 温湿度传感器
type TempHumiditySensor struct {
	IsChildInstance *IsChildInstance
	Name            *attribute.Name
	Model        	*attribute.Model
	Manufacturer 	*attribute.Manufacturer
	Version      	*attribute.Version

	Temperature *Temperature
	Humidity *Humidity
	Battery *Battery
}

func (w TempHumiditySensor) InstanceName() string {
	return "temperature_and_humidity_sensor"
}

// Temperature 温度
type Temperature struct {
	attribute.Int
}

func NewTemperature() *Temperature {
	return &Temperature{}
}

// Humidity 湿度
type Humidity struct {
	attribute.Int
}

func NewHumidity() *Humidity {
	return &Humidity{}
}