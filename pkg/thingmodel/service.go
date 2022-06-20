package thingmodel

type ServiceType string

const (
	InfoService               ServiceType = "info"               // 详情
	GatewayService            ServiceType = "gateway"            // 网关
	LightBulbService          ServiceType = "light_bulb"         // 灯
	SwitchService             ServiceType = "switch"             // 开关
	OutletService             ServiceType = "outlet"             // 插座
	CurtainService            ServiceType = "curtain"            // 窗帘电机 TODO 后续拆分电机
	TemperatureSensor         ServiceType = "temperature_sensor" // 温度
	HumiditySensor            ServiceType = "humidity_sensor"    // 湿度
	HeaterCooler              ServiceType = "heater_cooler"      // 加热器冷却器
	Lock                      ServiceType = "lock"
	Door                      ServiceType = "door"
	Doorbell                  ServiceType = "doorbell"
	MotionSensor              ServiceType = "motion_sensor"                // 人体传感器
	LeakSensor                ServiceType = "leak_sensor"                  // 水浸传感器
	BatteryService            ServiceType = "battery"                      // 电池服务
	SecuritySystem            ServiceType = "security_system"              // 安全服务
	StateLessSwitch           ServiceType = "stateless_switch"             // 无线开关
	ContactSensor             ServiceType = "contact_sensor"               // 接触式传感器
	Speaker                   ServiceType = "speaker"                      // 扬声器
	Microphone                ServiceType = "microphone"                   // 麦克风
	LightSensor               ServiceType = "light_sensor"                 // 光传感器
	CameraRTPStreamManagement ServiceType = "camera_rtp_stream_management" // 流管理服务
	OperatingMode             ServiceType = "operating_mode"               // 工作模式
	MediaNegotiation          ServiceType = "media_negotiation"            // webrtc媒体交换
	PTZ                       ServiceType = "ptz"                          // PTZ 摄像头云台控制功能
	Media                     ServiceType = "media"                        // Media 摄像头视频相关配置
)

type Service struct {
	Type       ServiceType `json:"type"`
	Attributes []Attribute `json:"attributes"`
}
