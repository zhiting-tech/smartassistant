package thingmodel

type ServiceType string

const (
	InfoService       ServiceType = "info"       // 详情
	GatewayService    ServiceType = "gateway"    // 网关
	LightBulbService  ServiceType = "light_bulb" // 灯
	SwitchService     ServiceType = "switch"     // 开关
	OutletService     ServiceType = "outlet"     // 插座
	CurtainService    ServiceType = "curtain"    // 窗帘电机 TODO 后续差分电机
	TemperatureSensor ServiceType = "temperature_sensor"
	HumiditySensor    ServiceType = "humidity_sensor"
	HeaterCooler      ServiceType = "heater_cooler"
	Lock              ServiceType = "lock"
	MotionSensor      ServiceType = "motion_sensor"
	LeakSensor        ServiceType = "leak_sensor"
	BatteryService    ServiceType = "battery"
	SecuritySystem    ServiceType = "security_system"
	StateLessSwitch   ServiceType = "stateless_switch"
	ContactSensor     ServiceType = "contact_sensor"
	Speaker           ServiceType = "speaker"
	Microphone        ServiceType = "microphone"
	LightSensor       ServiceType = "light_sensor"
)

type Service struct {
	Type       ServiceType `json:"type"`
	Attributes []Attribute `json:"attributes"`
}
