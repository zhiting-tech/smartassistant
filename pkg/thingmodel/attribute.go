package thingmodel

import (
	"encoding/json"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type ValType string

func (t ValType) String() string {
	return string(t)
}

const (
	Int     = ValType("int")
	Int32   = ValType("int32")
	Int64   = ValType("int64")
	String  = ValType("string")
	Bool    = ValType("bool")
	Float32 = ValType("float32")
	Float64 = ValType("float64")
	Enum    = ValType("enum") // number or string
	JSON    = ValType("json")
)

type Permission uint

const (
	AttributePermissionRead Permission = 1 << iota
	AttributePermissionWrite
	AttributePermissionNotify
	AttributePermissionHidden
	AttributePermissionSceneHidden
)

func SetPermissions(prs ...Permission) uint {
	var p uint
	for _, pr := range prs {
		p = p | uint(pr)
	}
	return p
}

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
	Default interface{} `json:"default,omitempty"`
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
	case Int, Int64, Int32, Float64, Float32:
		switch v := t.Val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		default:
			logger.Warnf("invalid val type(%s) of %s", v, t.Type)
		}
	default:
		logger.Warnf("invalid val type(%s) of %s", t.ValType, t.Type)
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

func (t Attribute) PermissionSceneHidden() bool {
	return IsContainPermissions(t.Permission, AttributePermissionSceneHidden)
}

func (t Attribute) NoPermission() bool {
	return t.Permission == 0
}

func (t *Attribute) RemovePermissions(prs ...Permission) {
	for _, pr := range prs {
		t.Permission = t.Permission ^ (uint(pr) & t.Permission)
	}
}

func (t *Attribute) SetPermissions(prs ...Permission) {
	t.Permission = SetPermissions(prs...)
}

type IAttribute interface {
	Set(interface{}) error
}

// PeerConnectionReq 媒体协商（offer和命令）
type PeerConnectionReq struct {
	Id             int64  `json:"id"`              // 唯一标识
	Offer          string `json:"offer"`           // RemoteDescription 远端SDP描述信息
	SessionCommand string `json:"session_command"` // 流会话控制命令(start：开启 end: 关闭 suspend:暂停 resume:继续播放 reconfigure:重新配置)
}

// PeerConnectionResp answer响应
type PeerConnectionResp struct {
	Id     int64  `json:"id"`     // 与PeerConnectionReq的id一致
	Answer string `json:"answer"` // LocalDescription 本地SDP描述信息
}

const (
	SubTypeBle = "ble" // 蓝牙
	SubTypeZb  = "zb"  // zigbee
)

// SubType 子设备类型
var SubType = Attribute{
	Type:    "sub_type",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// Volume 音量
var Volume = Attribute{
	Type:    "volume",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// OnOff 开关
var OnOff = Attribute{
	Type:    "on_off",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Brightness 亮度
var Brightness = Attribute{
	Type:    "brightness",
	ValType: Int32,
	Min:     1,
	Max:     100,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// ColorTemperature 色温
var ColorTemperature = Attribute{
	Type:    "color_temp",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// RGB RGB
var RGB = Attribute{
	Type:    "rgb",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Model 型号
var Model = Attribute{
	Type:    "model",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// Manufacturer 厂商
var Manufacturer = Attribute{
	Type:    "manufacturer",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// Identify 唯一标识
var Identify = Attribute{
	Type:    "identify",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// Version 固件版本
var Version = Attribute{
	Type:    "version",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// Name 设备名称
var Name = Attribute{
	Type:    "name",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

var Type = Attribute{
	Type:    "type",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// CurrentPosition 当前位置
var CurrentPosition = Attribute{
	Type:    "current_position",
	ValType: Int32,
	Min:     1,
	Max:     100,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// TargetPosition 目标位置
var TargetPosition = Attribute{
	Type:    "target_position",
	ValType: Int32,
	Min:     1,
	Max:     100,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// State 状态
var State = Attribute{
	Type:    "state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Direction 方向
var Direction = Attribute{
	Type:    "direction",
	ValType: Bool,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Humidity 湿度
var Humidity = Attribute{
	Type:    "humidity",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// Temperature 温度
var Temperature = Attribute{
	Type:    "temperature",
	ValType: Float32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// LeakDetected 泄漏检测
var LeakDetected = Attribute{
	Type:    "leak_detected",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// SwitchEvent 开关事件, 0: 单机; 1: 双击; 2: 长按
var SwitchEvent = Attribute{
	Type:    "switch_event",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// TargetState 目标状态
var TargetState = Attribute{
	Type:    "target_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// CurrentState 当前状态
var CurrentState = Attribute{
	Type:    "current_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// MotionDetected 移动检测
var MotionDetected = Attribute{
	Type:    "motion_detected",
	ValType: Bool,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// Battery 电池
var Battery = Attribute{
	Type:    "battery",
	ValType: Float32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// LockCurrentState 锁当前状态 //
var LockCurrentState = Attribute{
	Type:    "lock_current_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// LockTargetState 锁目标状态
var LockTargetState = Attribute{
	Type:    "lock_target_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

const (
	OpenSuccess = 1 + iota
	MultiOpenFail
	Ring
	UserAdd
	UserDel
	Pry
	Coerce
)

// LockEvent 锁事件 //1开门成功2多次验证失败3有人按门铃4新增用户5删除用户6撬锁告警7胁迫开门
var LockEvent = Attribute{
	Type:    "lock_event",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionNotify,
	),
}

// LockNotificationStructure 锁通知结构
type LockNotificationStructure struct {
	Time       int64  `json:"time"`
	EventType  int    `json:"event_type"`  // LockEvent //1开门成功2多次验证失败3有人按门铃4新增用户5删除用户6撬锁告警7胁迫开门
	NumberType int    `json:"number_type"` // 1指纹/2密码/6存储卡，开门成功/添加用户/删除用户/胁迫开门时有值
	NumberId   int    `json:"number_id"`   // 门锁本地编号，开门成功/添加用户/删除用户/胁迫开门时有值
	Username   string `json:"username"`    // 用户名，开门成功/胁迫开门时有值
}

// LockNotification 锁通知 json结构详见 LockNotificationStructure
var LockNotification = Attribute{
	Type:    "lock_notification",
	ValType: JSON,
	Permission: SetPermissions(
		AttributePermissionNotify,
	),
}

// Logs 日志
var Logs = Attribute{
	Type:    "logs",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Active 活动状态
var Active = Attribute{
	Type:    "active",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// CurrentTemperature 当前温度
var CurrentTemperature = Attribute{
	Type:    "current_temperature",
	ValType: Float32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// CurrentHeatingCoolingState 当前加热冷却状态
var CurrentHeatingCoolingState = Attribute{
	Type:    "current_heating_cooling_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// TargetHeatingCoolingState 目标加热冷却状态
var TargetHeatingCoolingState = Attribute{
	Type:    "target_heating_cooling_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// HeatingThresholdTemperature 加热阈值温度
var HeatingThresholdTemperature = Attribute{
	Type:    "heating_threshold_temperature",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// CoolingThresholdTemperature 冷却阈值温度
var CoolingThresholdTemperature = Attribute{
	Type:    "cooling_threshold_temperature",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// RotationSpeed 转速
var RotationSpeed = Attribute{
	Type:    "rotation_speed",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// SwingMode 摆动模式
var SwingMode = Attribute{
	Type:    "swing_mode",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// PermitJoin 是否允许加入
var PermitJoin = Attribute{
	Type:    "permit_join",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
		AttributePermissionHidden,
	),
}

// Alert 告警
var Alert = Attribute{
	Type:    "alert",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionNotify,
	),
}

// StatusLowBattery 低电量状态
var StatusLowBattery = Attribute{
	Type:    "status_low_battery",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// ContactSensorState 触点传感器状态
var ContactSensorState = Attribute{
	Type:    "contact_sensor_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

var Mute = Attribute{
	Type:    "mute",
	ValType: Bool,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

var CurrentAmbientLightLevel = Attribute{
	Type:    "current_ambient_light_level",
	ValType: Float32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// NightVision 夜视灯
var NightVision = Attribute{
	Type:    "night_vision",
	ValType: Bool,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

var ModeIndicator = Attribute{
	Type:    "mode_indicator",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

var WebRtcControl = Attribute{
	Type:    "webrtc_control",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionSceneHidden,
	),
}

var Answer = Attribute{
	Type:    "answer",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
		AttributePermissionHidden,
	),
}

// StreamingStatus 流状态，0：Available（空闲） 1： In Use（正在使用中） 2：Unavailable（不可用）
var StreamingStatus = Attribute{
	Type:    "streaming_status",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
		AttributePermissionHidden,
	),
}

// PTZMove 摄像头云台持续移动：{"pan":1,"tilt":0,"zoom":0} pan（水平移动速度:0.0-1.0） tilt（垂直移动速度:0.0-1.0）zoom：放大缩小
var PTZMove = Attribute{
	Type:    "move",
	ValType: String, // PTZMoveParam json
	Permission: SetPermissions(
		AttributePermissionWrite,
	),
}

// MediaResolutionOptions 摄像头分辨率可设属性
var MediaResolutionOptions = Attribute{
	Type:    "resolution_options",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// MediaResolution 摄像头分辨率属性
var MediaResolution = Attribute{
	Type:    "resolution",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
		AttributePermissionSceneHidden,
	),
}

// MediaFrameRateLimit 摄像头帧率
var MediaFrameRateLimit = Attribute{
	Type:    "frame_rate_limit",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
		AttributePermissionHidden,
	)}

// MediaBitRateLimit 摄像头码率
var MediaBitRateLimit = Attribute{
	Type:    "bitrate_limit",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
		AttributePermissionHidden,
	),
}

// MediaEncodingInterval 摄像头编码区间
var MediaEncodingInterval = Attribute{
	Type:    "encoding_interval",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
		AttributePermissionHidden,
	),
}

// MediaQuality 摄像头视频质量
var MediaQuality = Attribute{
	Type:    "media_quality",
	ValType: Float32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
		AttributePermissionHidden,
	),
}

// MediaGovLength 摄像头I帧间隔长度
var MediaGovLength = Attribute{
	Type:    "gov_length",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
		AttributePermissionHidden,
	),
}

// PTZTDCruise 摄像头云台上下巡航 （开始：td_start 停止：td_stop）
var PTZTDCruise = Attribute{
	Type:    "top_down_cruise",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// PTZLRCruise 摄像头云台左右巡航 （开始：lr_start 停止：lr_stop）
var PTZLRCruise = Attribute{
	Type:    "left_right_cruise",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Select SelectItems的选择结构
type Select struct {
	ID    *int64      `json:"id,omitempty"`
	Items []SelectItem `json:"items,omitempty"`
}

type SelectItem struct {
	ID   *int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func NewSelectAttr() Select {
	var defaultID int64 = 0
	return Select{
		ID: &defaultID,
		Items: make([]SelectItem, 0),
	}
}

func (s *Select) SetDefaultID(ID int64) {
	s.ID = &ID
}

func (s *Select) GetDefaultID() int64{
	if s.ID == nil {
		return 0
	}
	return *s.ID
}

func (s *Select) Add(item SelectItem) {
	s.Items = append(s.Items, item)
}

func (s *Select) Remove(targetItem SelectItem) {
	s.ForEachItems(func(index int, item SelectItem) bool {
		if targetItem.ID == item.ID {
			s.Items = append(s.Items[:index], s.Items[index+1:]...)
			return false
		}
		return true
	})
}

func (s *Select) Marshal() (string, error){
	jsonData, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (s *Select) ForEachItems(f func(index int, item SelectItem) bool) {
	for i, it := range s.Items {
		if ok := f(i, it); !ok {
			return
		}
	}
}

// SelectItems 可用于sa场景选择设置
var SelectItems = Attribute{
	Type:    "select_items",
	ValType: JSON,
	Permission: SetPermissions(
		AttributePermissionWrite,
	),
}

func SelectUnmarshal(data []byte) (result Select, err error) {
	err = json.Unmarshal(data, &result)
	return
}