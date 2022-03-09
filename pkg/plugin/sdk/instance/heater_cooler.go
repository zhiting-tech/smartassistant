package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

// HeaterCooler 加热冷却
type HeaterCooler struct {
	Active                      *Active
	CurrentTemperature          *Temperature       // 当前温度
	CoolingThresholdTemperature *Temperature       // 冷却阈值温度
	HeatingThresholdTemperature *Temperature       // 加热阈值温度
	CurrentHeaterCoolerState    *HeaterCoolerState // 当前加热器冷却器状态，1 空闲，2 加热，3 冷却
	TargetHeaterCoolerState     *HeaterCoolerState // 当前加热器冷却器状态，1 加热，2 冷却
	RotationSpeed               *RotationSpeed     // 风扇旋转速度，0-100，step 25
	SwingMode                   *SwingMode         // 摆动模式，0 不摆动，1 摆动
}

// Active 活动状态,0 非活动,1 活动
type Active struct {
	attribute.Int
}

func NewActive() *Active {
	return &Active{}
}

// HeaterCoolerState 加热冷却器的当前状态 0“无效”“1”空闲”“2”加热“3”冷却
type HeaterCoolerState struct {
	attribute.Int
}

func NewHeaterCoolerState() *HeaterCoolerState {
	return &HeaterCoolerState{}
}

// RotationSpeed 风扇旋转速度
type RotationSpeed struct {
	attribute.Int
}

func NewRotationSpeed() *RotationSpeed {
	return &RotationSpeed{}
}

// SwingMode 摆动模式，0 不摆动，1 摆动
type SwingMode struct {
	attribute.Int
}

func NewSwingMode() *SwingMode {
	return &SwingMode{}
}
