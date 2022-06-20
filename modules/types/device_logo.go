package types

import (
	"net/http"

	"github.com/zhiting-tech/smartassistant/modules/utils/url"
)

type DeviceLogo struct {
	LogoType LogoType
	Name     string
	FileName string
}

type LogoType int

const (
	NormalLogo LogoType = iota
	LightLogo
	SwitchLogo
	OutletLogo
	CurtainLogo
	GatewayLogo
	SensorLogo
	CameraLogo
	DoorLockLogo
	OthLogo
	TemperatureAndHumiditySensorLogo
	WaterLeakSensorLogo
	HumanSensorLogo
	WindowDoorSensorLogo
	SmokeSensorLogo
	GasSensorLogo
)

var (
	DeviceLogos = []DeviceLogo{
		{LogoType: LightLogo, Name: "灯", FileName: "light.png"},
		{LogoType: SwitchLogo, Name: "开关", FileName: "switch.png"},
		{LogoType: OutletLogo, Name: "插座", FileName: "outlet.png"},
		{LogoType: CurtainLogo, Name: "窗帘电机", FileName: "curtain.png"},
		{LogoType: GatewayLogo, Name: "网关", FileName: "gateway.png"},
		{LogoType: CameraLogo, Name: "摄像头", FileName: "camera.png"},
		{LogoType: DoorLockLogo, Name: "智能门锁", FileName: "door_lock.png"},
		{LogoType: SensorLogo, Name: "传感器", FileName: "sensor.png"},
		{LogoType: TemperatureAndHumiditySensorLogo, Name: "温湿度传感器", FileName: "temperature_humidity_sensor.png"},
		{LogoType: WaterLeakSensorLogo, Name: "水浸传感器", FileName: "water_leak_sensor.png"},
		{LogoType: HumanSensorLogo, Name: "人体传感器", FileName: "human_sensor.png"},
		{LogoType: WindowDoorSensorLogo, Name: "门磁传感器", FileName: "window_door_sensor.png"},
		{LogoType: SmokeSensorLogo, Name: "烟雾报警器", FileName: "smoke_sensor.png"},
		{LogoType: GasSensorLogo, Name: "燃气报警器", FileName: "gas_sensor.png"},
		{LogoType: OthLogo, Name: "其他", FileName: "other_device.png"},
	}

	NormalLogoInfo = DeviceLogo{
		LogoType: NormalLogo,
		Name:     "设备图片",
	}
)

func LogoFromLogoType(logoType LogoType, httpReq *http.Request) (logoUrl, logo string) {

	if logoInfo, ok := GetLogo(logoType); ok {
		logoUrl = url.ImageUrl(httpReq, logoInfo.FileName)
		logo = url.ImagePath(logoInfo.FileName)
	}
	return
}

func GetLogo(logoType LogoType) (DeviceLogo, bool) {
	for _, l := range DeviceLogos {
		if l.LogoType == logoType {
			return l, true
		}
	}
	return DeviceLogo{}, false
}
