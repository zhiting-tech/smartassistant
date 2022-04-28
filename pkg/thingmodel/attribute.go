package thingmodel

import "github.com/zhiting-tech/smartassistant/pkg/logger"

type ValType string

func (t ValType) String() string {
	return string(t)
}

const (
	Int32   = ValType("int32")
	Int64   = ValType("int64")
	String  = ValType("string")
	Bool    = ValType("bool")
	Float32 = ValType("float32")
	Float64 = ValType("float64")
	Enum    = ValType("enum") // number or string
)

type Permission uint

const (
	AttributePermissionRead Permission = 1 << iota
	AttributePermissionWrite
	AttributePermissionNotify
	AttributePermissionHidden
)

func IsContainPermissions(permission uint, targets ...Permission) bool {
	for _, t := range targets {
		if t != Permission(permission)&t {
			return false
		}
	}
	return true
}

type Attribute struct {
	AID        int    `json:"aid"`
	Type       string `json:"type"`
	Permission uint   `json:"permission"`

	ValType ValType     `json:"val_type"`
	Val     interface{} `json:"val"`
	Min     interface{} `json:"min,omitempty"`
	Max     interface{} `json:"max,omitempty"`
}

func (t Attribute) GetString() string {
	if t.ValType == String {
		val, _ := t.Val.(string)
		return val
	}
	return ""
}

func (t Attribute) GetInt() int {
	switch t.ValType {
	case Int64, Int32, Float64, Float32:
		switch v := t.Val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		default:
			logger.Warnf("invalid val type(%s) of %s", v, t.Type)
		}
	}
	return 0
}

func (t Attribute) String() string {
	return t.Type
}

func (t Attribute) PermissionWrite() bool {
	return IsContainPermissions(t.Permission, AttributePermissionWrite)
}
func (t Attribute) PermissionRead() bool {
	return IsContainPermissions(t.Permission, AttributePermissionRead)
}

func (t Attribute) PermissionNotify() bool {
	return IsContainPermissions(t.Permission, AttributePermissionNotify)
}

func (t Attribute) PermissionHidden() bool {
	return IsContainPermissions(t.Permission, AttributePermissionHidden)
}

func (t Attribute) NoPermission() bool {
	return t.Permission == 0
}

func (t *Attribute) SetPermissions(prs ...Permission) {
	var p uint
	for _, pr := range prs {
		p = p | uint(pr)
	}
	t.Permission = p
}

type IAttribute interface {
	Set(interface{}) error
}

// Volume 音量
var Volume = Attribute{
	Type:    "volume",
	ValType: Int32,
}

// OnOff 开关
var OnOff = Attribute{
	Type:    "on_off",
	ValType: String,
}

// Brightness 亮度
var Brightness = Attribute{
	Type:    "brightness",
	ValType: Int32,
	Min:     1,
	Max:     100,
}

// ColorTemperature 色温
var ColorTemperature = Attribute{
	Type:    "color_temp",
	ValType: Int32,
}

// RGB RGB
var RGB = Attribute{
	Type:    "rgb",
	ValType: String,
}

// Model 型号
var Model = Attribute{
	Type:    "model",
	ValType: String,
}

// Manufacturer 厂商
var Manufacturer = Attribute{
	Type:    "manufacturer",
	ValType: String,
}

// Identify 唯一标识
var Identify = Attribute{
	Type:    "identify",
	ValType: String,
}

// Version 固件版本
var Version = Attribute{
	Type:    "version",
	ValType: String,
}

// Name 设备名称
var Name = Attribute{
	Type:    "name",
	ValType: String,
}

// CurrentPosition 当前位置
var CurrentPosition = Attribute{
	Type:    "current_position",
	ValType: Int32,
	Min:     1,
	Max:     100,
}

// TargetPosition 目标位置
var TargetPosition = Attribute{
	Type:    "target_position",
	ValType: Int32,
	Min:     1,
	Max:     100,
}

// State 状态
var State = Attribute{
	Type:    "state",
	ValType: Int32,
}

// Direction 方向
var Direction = Attribute{
	Type:    "direction",
	ValType: Bool,
}

// Humidity 湿度
var Humidity = Attribute{
	Type:    "humidity",
	ValType: Int32,
}

// Temperature 温度
var Temperature = Attribute{
	Type:    "temperature",
	ValType: Float32,
}

// LeakDetected 泄漏检测
var LeakDetected = Attribute{
	Type:    "leak_detected",
	ValType: Int32,
}

// SwitchEvent 开关事件
var SwitchEvent = Attribute{
	Type:    "switch_event",
	ValType: Int32,
}

// TargetState 目标状态
var TargetState = Attribute{
	Type:    "target_state",
	ValType: Int32,
}

// CurrentState 当前状态
var CurrentState = Attribute{
	Type:    "current_state",
	ValType: Int32,
}

// MotionDetected 移动检测
var MotionDetected = Attribute{
	Type:    "motion_detected",
	ValType: Bool,
}

// Battery 电池
var Battery = Attribute{
	Type:    "battery",
	ValType: Float32,
}

// LockTargetState 锁目标状态
var LockTargetState = Attribute{
	Type:    "lock_target_state",
	ValType: Int32,
}

// Logs 日志
var Logs = Attribute{
	Type:    "logs",
	ValType: String,
}

// Active 活动状态
var Active = Attribute{
	Type:    "active",
	ValType: Int32,
}

// CurrentTemperature 当前温度
var CurrentTemperature = Attribute{
	Type:    "current_temperature",
	ValType: Float32,
}

// CurrentHeatingCoolingState 当前加热冷却状态
var CurrentHeatingCoolingState = Attribute{
	Type:    "current_heating_cooling_state",
	ValType: Int32,
}

// TargetHeatingCoolingState 目标加热冷却状态
var TargetHeatingCoolingState = Attribute{
	Type:    "target_heating_cooling_state",
	ValType: Int32,
}

// HeatingThresholdTemperature 加热阈值温度
var HeatingThresholdTemperature = Attribute{
	Type:    "heating_threshold_temperature",
	ValType: Int32,
}

// CoolingThresholdTemperature 冷却阈值温度
var CoolingThresholdTemperature = Attribute{
	Type:    "cooling_threshold_temperature",
	ValType: Int32,
}

// RotationSpeed 转速
var RotationSpeed = Attribute{
	Type:    "rotation_speed",
	ValType: Int32,
}

// SwingMode 摆动模式
var SwingMode = Attribute{
	Type:    "swing_mode",
	ValType: Int32,
}

// PermitJoin 是否允许加入
var PermitJoin = Attribute{
	Type:    "permit_join",
	ValType: Int32,
}

// Alert 告警
var Alert = Attribute{
	Type:       "alert",
	ValType:    Int32,
	Permission: 0,
}

// StatusLowBattery 低电量状态
var StatusLowBattery = Attribute{
	Type:    "status_low_battery",
	ValType: Int32,
}

// ContactSensorState 触点传感器状态
var ContactSensorState = Attribute{
	Type:    "contact_sensor_state",
	ValType: Int32,
}

var Mute = Attribute{
	Type:    "mute",
	ValType: Bool,
}

var CurrentAmbientLightLevel = Attribute{
	Type:    "current_ambient_light_level",
	ValType: Float32,
}

// init 初始化设备权限
func init() {
	Volume.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	OnOff.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	Brightness.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	ColorTemperature.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	RGB.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	Manufacturer.SetPermissions(
		AttributePermissionRead,
	)
	Identify.SetPermissions(
		AttributePermissionRead,
	)
	Model.SetPermissions(
		AttributePermissionRead,
	)
	Name.SetPermissions(
		AttributePermissionRead,
	)
	Version.SetPermissions(
		AttributePermissionRead,
	)
	CurrentPosition.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	TargetPosition.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	State.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	Direction.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	Humidity.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	Temperature.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	LeakDetected.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	SwitchEvent.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	TargetState.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	CurrentState.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	MotionDetected.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	Battery.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	LockTargetState.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	Logs.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	Active.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	CurrentTemperature.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	CurrentHeatingCoolingState.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	TargetHeatingCoolingState.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	RotationSpeed.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	SwingMode.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	PermitJoin.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
		AttributePermissionHidden,
	)
	StatusLowBattery.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	ContactSensorState.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
	Mute.SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	)
	CurrentAmbientLightLevel.SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	)
}
