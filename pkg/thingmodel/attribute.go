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

// PeerConnectionReq ???????????????offer????????????
type PeerConnectionReq struct {
	Id             int64  `json:"id"`              // ????????????
	Offer          string `json:"offer"`           // RemoteDescription ??????SDP????????????
	SessionCommand string `json:"session_command"` // ?????????????????????(start????????? end: ?????? suspend:?????? resume:???????????? reconfigure:????????????)
}

// PeerConnectionResp answer??????
type PeerConnectionResp struct {
	Id     int64  `json:"id"`     // ???PeerConnectionReq???id??????
	Answer string `json:"answer"` // LocalDescription ??????SDP????????????
}

const (
	SubTypeBle = "ble" // ??????
	SubTypeZb  = "zb"  // zigbee
)

// SubType ???????????????
var SubType = Attribute{
	Type:    "sub_type",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// Volume ??????
var Volume = Attribute{
	Type:    "volume",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// OnOff ??????
var OnOff = Attribute{
	Type:    "on_off",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Brightness ??????
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

// ColorTemperature ??????
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

// Model ??????
var Model = Attribute{
	Type:    "model",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// Manufacturer ??????
var Manufacturer = Attribute{
	Type:    "manufacturer",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// Identify ????????????
var Identify = Attribute{
	Type:    "identify",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// Version ????????????
var Version = Attribute{
	Type:    "version",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// Name ????????????
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

// CurrentPosition ????????????
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

// TargetPosition ????????????
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

// State ??????
var State = Attribute{
	Type:    "state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Direction ??????
var Direction = Attribute{
	Type:    "direction",
	ValType: Bool,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Humidity ??????
var Humidity = Attribute{
	Type:    "humidity",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// Temperature ??????
var Temperature = Attribute{
	Type:    "temperature",
	ValType: Float32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// LeakDetected ????????????
var LeakDetected = Attribute{
	Type:    "leak_detected",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// SwitchEvent ????????????, 0: ??????; 1: ??????; 2: ??????
var SwitchEvent = Attribute{
	Type:    "switch_event",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// TargetState ????????????
var TargetState = Attribute{
	Type:    "target_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// CurrentState ????????????
var CurrentState = Attribute{
	Type:    "current_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// MotionDetected ????????????
var MotionDetected = Attribute{
	Type:    "motion_detected",
	ValType: Bool,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// Battery ??????
var Battery = Attribute{
	Type:    "battery",
	ValType: Float32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// LockCurrentState ??????????????? //
var LockCurrentState = Attribute{
	Type:    "lock_current_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// LockTargetState ???????????????
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

// LockEvent ????????? //1????????????2??????????????????3???????????????4????????????5????????????6????????????7????????????
var LockEvent = Attribute{
	Type:    "lock_event",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionNotify,
	),
}

// LockNotificationStructure ???????????????
type LockNotificationStructure struct {
	Time       int64  `json:"time"`
	EventType  int    `json:"event_type"`  // LockEvent //1????????????2??????????????????3???????????????4????????????5????????????6????????????7????????????
	NumberType int    `json:"number_type"` // 1??????/2??????/6????????????????????????/????????????/????????????/?????????????????????
	NumberId   int    `json:"number_id"`   // ?????????????????????????????????/????????????/????????????/?????????????????????
	Username   string `json:"username"`    // ????????????????????????/?????????????????????
}

// LockNotification ????????? json???????????? LockNotificationStructure
var LockNotification = Attribute{
	Type:    "lock_notification",
	ValType: JSON,
	Permission: SetPermissions(
		AttributePermissionNotify,
	),
}

// Logs ??????
var Logs = Attribute{
	Type:    "logs",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Active ????????????
var Active = Attribute{
	Type:    "active",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// CurrentTemperature ????????????
var CurrentTemperature = Attribute{
	Type:    "current_temperature",
	ValType: Float32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// CurrentHeatingCoolingState ????????????????????????
var CurrentHeatingCoolingState = Attribute{
	Type:    "current_heating_cooling_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// TargetHeatingCoolingState ????????????????????????
var TargetHeatingCoolingState = Attribute{
	Type:    "target_heating_cooling_state",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// HeatingThresholdTemperature ??????????????????
var HeatingThresholdTemperature = Attribute{
	Type:    "heating_threshold_temperature",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// CoolingThresholdTemperature ??????????????????
var CoolingThresholdTemperature = Attribute{
	Type:    "cooling_threshold_temperature",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// RotationSpeed ??????
var RotationSpeed = Attribute{
	Type:    "rotation_speed",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// SwingMode ????????????
var SwingMode = Attribute{
	Type:    "swing_mode",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// PermitJoin ??????????????????
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

// Alert ??????
var Alert = Attribute{
	Type:    "alert",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionNotify,
	),
}

// StatusLowBattery ???????????????
var StatusLowBattery = Attribute{
	Type:    "status_low_battery",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
	),
}

// ContactSensorState ?????????????????????
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

// NightVision ?????????
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

// StreamingStatus ????????????0???Available???????????? 1??? In Use????????????????????? 2???Unavailable???????????????
var StreamingStatus = Attribute{
	Type:    "streaming_status",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionNotify,
		AttributePermissionHidden,
	),
}

// PTZMove ??????????????????????????????{"pan":1,"tilt":0,"zoom":0} pan?????????????????????:0.0-1.0??? tilt?????????????????????:0.0-1.0???zoom???????????????
var PTZMove = Attribute{
	Type:    "move",
	ValType: String, // PTZMoveParam json
	Permission: SetPermissions(
		AttributePermissionWrite,
	),
}

// MediaResolutionOptions ??????????????????????????????
var MediaResolutionOptions = Attribute{
	Type:    "resolution_options",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
	),
}

// MediaResolution ????????????????????????
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

// MediaFrameRateLimit ???????????????
var MediaFrameRateLimit = Attribute{
	Type:    "frame_rate_limit",
	ValType: Int32,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
		AttributePermissionHidden,
	)}

// MediaBitRateLimit ???????????????
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

// MediaEncodingInterval ?????????????????????
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

// MediaQuality ?????????????????????
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

// MediaGovLength ?????????I???????????????
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

// PTZTDCruise ??????????????????????????? ????????????td_start ?????????td_stop???
var PTZTDCruise = Attribute{
	Type:    "top_down_cruise",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// PTZLRCruise ??????????????????????????? ????????????lr_start ?????????lr_stop???
var PTZLRCruise = Attribute{
	Type:    "left_right_cruise",
	ValType: String,
	Permission: SetPermissions(
		AttributePermissionRead,
		AttributePermissionWrite,
		AttributePermissionNotify,
	),
}

// Select SelectItems???????????????
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

// SelectItems ?????????sa??????????????????
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