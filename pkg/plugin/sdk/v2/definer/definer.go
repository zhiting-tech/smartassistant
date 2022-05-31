package definer

import (
	"encoding/json"
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

var NotEnableErr = errors.New("attribute not enable")
var NotFoundErr = errors.New("attribute not found")

type NotifyFunc func(event AttributeEvent) error
type ThingModelNotifyFunc func(iid string, event ThingModelEvent) error

type SetRequest struct {
	IID   string
	AID   int
	Value interface{}
}
type GetRequest struct {
	IID string
	AID int
}

func NewThingModelDefiner(iid string, fn NotifyFunc, tfn ThingModelNotifyFunc) *Definer {
	t := Definer{
		iid:                  iid,
		instanceMap:          make(map[string]*Instance),
		thingModelNotifyFunc: tfn,
	}
	t.notifyFunc = func(event AttributeEvent) error {
		// 子设备的更新需要同时更新父设备
		if event.IID != t.iid {
			t.UpdateThingModel()
		}
		return fn(event)
	}

	return &t
}

type Definer struct {
	iid         string
	instanceMap map[string]*Instance

	notifyFunc           NotifyFunc
	thingModelNotifyFunc ThingModelNotifyFunc
}

// Deprecated: 之后版本会删除，请使用baseService的Notify
func (t Definer) Notify(iid string, aid int, val interface{}) error {
	ev := AttributeEvent{
		IID: iid,
		AID: aid,
		Val: val,
	}

	// 子设备的更新需要同时更新父设备
	if iid != t.iid {
		t.UpdateThingModel()
	}

	return t.notifyFunc(ev)
}

func (t *Definer) Instance(iid string) *Instance {
	if i, ok := t.instanceMap[iid]; ok {
		return i
	}
	t.instanceMap[iid] = NewInstance(iid)
	return t.instanceMap[iid]
}

func (t *Definer) DelInstance(iid string) {
	delete(t.instanceMap, iid)
}

func (t Definer) UpdateThingModel() error {
	tm := t.ThingModel()

	tme := ThingModelEvent{
		ThingModel: tm,
		IID:        t.iid,
	}
	return t.thingModelNotifyFunc(t.iid, tme)
}

func (t *Definer) FromJSON(data []byte) {
	var tm thingmodel.ThingModel
	_ = json.Unmarshal(data, &tm)

	for _, ins := range tm.Instances {
		instance := t.Instance(ins.IID)
		for _, srv := range ins.Services {
			ss := instance.NewService(srv.Type)
			for _, a := range srv.Attributes {
				switch a.Type {
				case thingmodel.OnOff.Type:
					ss.WithAttribute(thingmodel.OnOff)
				case thingmodel.Brightness.Type:
					ss.WithAttribute(thingmodel.Brightness)
				case thingmodel.ColorTemperature.Type:
					ss.WithAttribute(thingmodel.ColorTemperature)
				case thingmodel.RGB.Type:
					ss.WithAttribute(thingmodel.RGB)
				case thingmodel.Model.Type:
					ss.WithAttribute(thingmodel.Model)
				case thingmodel.Manufacturer.Type:
					ss.WithAttribute(thingmodel.Manufacturer)
				case thingmodel.Identify.Type:
					ss.WithAttribute(thingmodel.Identify)
				case thingmodel.Version.Type:
					ss.WithAttribute(thingmodel.Version)
				case thingmodel.CurrentPosition.Type:
					ss.WithAttribute(thingmodel.CurrentPosition)
				case thingmodel.TargetPosition.Type:
					ss.WithAttribute(thingmodel.CurrentPosition)
				case thingmodel.State.Type:
					ss.WithAttribute(thingmodel.State)
				case thingmodel.Direction.Type:
					ss.WithAttribute(thingmodel.Direction)
				case thingmodel.Humidity.Type:
					ss.WithAttribute(thingmodel.Humidity)
				case thingmodel.Temperature.Type:
					ss.WithAttribute(thingmodel.Humidity)
				case thingmodel.LeakDetected.Type:
					ss.WithAttribute(thingmodel.LeakDetected)
				case thingmodel.SwitchEvent.Type:
					ss.WithAttribute(thingmodel.SwitchEvent)
				case thingmodel.TargetState.Type:
					ss.WithAttribute(thingmodel.TargetState)
				case thingmodel.CurrentState.Type:
					ss.WithAttribute(thingmodel.CurrentState)
				case thingmodel.MotionDetected.Type:
					ss.WithAttribute(thingmodel.MotionDetected)
				case thingmodel.Battery.Type:
					ss.WithAttribute(thingmodel.Battery)
				case thingmodel.LockTargetState.Type:
					ss.WithAttribute(thingmodel.LockTargetState)
				case thingmodel.Logs.Type:
					ss.WithAttribute(thingmodel.Logs)
				case thingmodel.Active.Type:
					ss.WithAttribute(thingmodel.Active)
				case thingmodel.CurrentTemperature.Type:
					ss.WithAttribute(thingmodel.CurrentTemperature)
				case thingmodel.CurrentHeatingCoolingState.Type:
					ss.WithAttribute(thingmodel.CurrentHeatingCoolingState)
				case thingmodel.TargetHeatingCoolingState.Type:
					ss.WithAttribute(thingmodel.TargetHeatingCoolingState)
				case thingmodel.RotationSpeed.Type:
					ss.WithAttribute(thingmodel.RotationSpeed)
				case thingmodel.SwingMode.Type:
					ss.WithAttribute(thingmodel.SwingMode)
				case thingmodel.PermitJoin.Type:
					ss.WithAttribute(thingmodel.PermitJoin)
				case thingmodel.StatusLowBattery.Type:
					ss.WithAttribute(thingmodel.StatusLowBattery)
				}

			}
		}
	}
	return
}

func (t *Definer) SetNotifyFunc() {
	logrus.Debug("definer set notify func")
	for _, i := range t.instanceMap {
		for _, a := range i.Services {
			a.SetNotifyFunc(i.IID, t.notifyFunc)
		}
	}
}

func (t *Definer) ThingModel() (tm thingmodel.ThingModel) {

	for _, i := range t.instanceMap { // TODO 丢失顺序，尝试优化
		ins := thingmodel.Instance{IID: i.IID}
		for _, s := range i.Services {
			srv := thingmodel.Service{Type: s.Type()}
			for _, a := range s.attributeMap {
				srv.Attributes = append(srv.Attributes, *a.meta)
			}
			ins.Services = append(ins.Services, srv)
		}
		if ins.IsGateway() {
			tm.Instances = append([]thingmodel.Instance{ins}, tm.Instances...)
		} else {
			tm.Instances = append(tm.Instances, ins)
		}
	}
	return
}

func (t Definer) SetAttribute(iid string, aid int, val interface{}) error {

	attr := t.getAttribute(iid, aid)
	if attr == nil {
		return NotFoundErr
	}
	return attr.Set(val)
}

func (t *Definer) getAttribute(iid string, aid int) *Attribute {
	ins := t.Instance(iid)
	return ins.GetAttribute(aid)
}

func NewInstance(id string) *Instance {
	ins := Instance{IID: id, attributes: make(map[int]*Attribute)}
	return &ins
}

type Instance struct {
	IID        string         `json:"iid"`
	Services   []*BaseService `json:"services"`
	attributes map[int]*Attribute
	i          int
}

func (t *Instance) AddAttribute(attr *Attribute) {
	t.i += 1
	attr.meta.AID = t.i
	t.attributes[t.i] = attr
}

func (t *Instance) Attributes() map[int]*Attribute {
	return t.attributes
}

func (t *Instance) GetAttribute(aid int) *Attribute {
	return t.attributes[aid]
}

func (t *Instance) NewService(serviceType thingmodel.ServiceType) *BaseService {
	srv := NewService(serviceType)
	srv._instance = t
	t.Services = append(t.Services, srv)
	return srv
}

func (t *Instance) NewInfo() *BaseService {
	return t.NewService(thingmodel.InfoService).
		WithAttributes(thingmodel.Manufacturer,
			thingmodel.Identify, thingmodel.Model, thingmodel.Version)
}

func (t *Instance) NewSwitch() *BaseService {
	return t.NewService(thingmodel.SwitchService).
		WithAttributes(thingmodel.OnOff)
}
func (t *Instance) NewOutlet() *BaseService {
	return t.NewService(thingmodel.OutletService).
		WithAttributes(thingmodel.OnOff)
}

func (t *Instance) NewLight() *BaseService {
	return t.NewService(thingmodel.LightBulbService).
		WithAttributes(
			thingmodel.OnOff,
			// thingmodel.Brightness,
			// thingmodel.ColorTemperature,
			// thingmodel.RGB,
		)
}

func (t *Instance) NewCurtain() *BaseService {
	return t.NewService(thingmodel.CurtainService).
		WithAttributes(
			thingmodel.CurrentPosition,
			thingmodel.TargetPosition,
			thingmodel.State,
			thingmodel.Direction,
		)
}
func (t *Instance) NewGateway() *BaseService {
	return t.NewService(thingmodel.GatewayService).
		WithAttributes(
		// thingmodel.PermitJoin,
		// thingmodel.Volume,
		)
}
func (t *Instance) NewHumiditySensor() *BaseService {
	return t.NewService(thingmodel.HumiditySensor).
		WithAttributes(thingmodel.Humidity)
}
func (t *Instance) NewTemperatureSensor() *BaseService {
	return t.NewService(thingmodel.TemperatureSensor).
		WithAttributes(thingmodel.Temperature)
}
func (t *Instance) NewHeaterCooler() *BaseService {
	return t.NewService(thingmodel.HeaterCooler).
		WithAttributes(
			thingmodel.Active,
			thingmodel.CurrentTemperature,
			thingmodel.CurrentHeatingCoolingState,
			thingmodel.TargetHeatingCoolingState,
			thingmodel.HeatingThresholdTemperature,
			thingmodel.CoolingThresholdTemperature)
}
func (t *Instance) NewLeakSensor() *BaseService {
	return t.NewService(thingmodel.LeakSensor).
		WithAttributes(thingmodel.LeakDetected)
}
func (t *Instance) NewLock() *BaseService {
	return t.NewService(thingmodel.Lock).
		WithAttributes(thingmodel.Battery)
}
func (t *Instance) NewDoor() *BaseService {
	return t.NewService(thingmodel.Door).
		WithAttributes(thingmodel.CurrentPosition)
}
func (t *Instance) NewDoorbell() *BaseService {
	return t.NewService(thingmodel.Doorbell).
		WithAttributes(thingmodel.SwitchEvent)
}
func (t *Instance) NewMotionSensor() *BaseService {
	return t.NewService(thingmodel.MotionSensor).
		WithAttributes(thingmodel.MotionDetected)
}

func (t *Instance) NewBatteryService() *BaseService {
	return t.NewService(thingmodel.BatteryService).
		WithAttributes(thingmodel.Battery)
}

func (t *Instance) NewSecuritySystem() *BaseService {
	return t.NewService(thingmodel.SecuritySystem).
		WithAttributes(thingmodel.CurrentState, thingmodel.TargetState)
}

func (t *Instance) NewStateLessSwitch() *BaseService {
	return t.NewService(thingmodel.StateLessSwitch).
		WithAttributes(thingmodel.SwitchEvent)
}

func (t *Instance) NewContactSensor() *BaseService {
	return t.NewService(thingmodel.ContactSensor).
		WithAttributes(thingmodel.ContactSensorState)
}

func (t *Instance) NewSpeaker() *BaseService {
	return t.NewService(thingmodel.Speaker).
		WithAttributes(
			thingmodel.Volume,
		// thingmodel.Mute
		)
}

func (t *Instance) NewMicrophone() *BaseService {
	return t.NewService(thingmodel.Microphone).
		WithAttributes(
			thingmodel.Volume,
		// thingmodel.Mute
		)
}

func (t *Instance) NewLightSensor() *BaseService {
	return t.NewService(thingmodel.LightSensor).
		WithAttributes(thingmodel.CurrentAmbientLightLevel)
}

// NewCameraRTPStreamManagement 摄像头RTP流服务
func (t *Instance) NewCameraRTPStreamManagement() *BaseService {
	return t.NewService(thingmodel.CameraRTPStreamManagement).
		WithAttributes(thingmodel.StreamingStatus)
}

// NewOperatingMode 设备工作模式服务
func (t *Instance) NewOperatingMode() *BaseService {
	return t.NewService(thingmodel.OperatingMode)
}

// NewMediaNegotiation webrtc媒体协商服务
func (t *Instance) NewMediaNegotiation() *BaseService {
	return t.NewService(thingmodel.MediaNegotiation).WithAttributes(
		thingmodel.WebRtcControl, thingmodel.Answer)
}

type Service struct {
	Type       thingmodel.ServiceType `json:"type"`
	Attributes []Attribute            `json:"attributes"`
}
