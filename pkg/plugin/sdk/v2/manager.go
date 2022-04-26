package sdk

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

var ThingModelNotFoundErr = errors.New("thing model not found")

func newDevice(d Device) *device {
	return &device{d, nil, atomic.NewBool(false)}
}

type device struct {
	Device
	df *definer.Definer

	connected *atomic.Bool
}

type Manager struct {
	discoveredDevices []Device
	devices           sync.Map // iid: device
	eventChannel      EventChan
	eventChannels     map[EventChan]struct{}
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Init() {
	m.eventChannel = make(EventChan, 10)
	m.eventChannels = make(map[EventChan]struct{})

	// 转发notifyChan消息到所有notifyChans
	go func() {
		for {
			select {
			case n := <-m.eventChannel:
				for ch := range m.eventChannels {
					select {
					case ch <- n:
					default:
					}
				}
			}
		}
	}()
}

func (m *Manager) Subscribe(ch EventChan) {
	m.eventChannels[ch] = struct{}{}
}

func (m *Manager) Unsubscribe(ch EventChan) {
	delete(m.eventChannels, ch)
}

// InitOrUpdateDevice 添加或更新设备
func (m *Manager) InitOrUpdateDevice(device Device) error {
	if device == nil {
		return errors.New("device is nil")
	}
	_, loaded := m.devices.LoadOrStore(device.Info().IID, newDevice(device))
	if loaded {
		logrus.Debugf("device %s already exist", device.Info().IID)
		return nil
	}
	logrus.Info("add device:", device.Info())

	return nil
}

func (m *Manager) IsOTASupport(iid string) (ok bool, err error) {

	d, err := m.getDevice(iid)
	if err != nil {
		return
	}

	switch d.(type) {
	case OTADevice:
		return true, nil
	default:
		return
	}
}

func (m *Manager) OTA(iid, firmwareURL string) (ch chan OTAResp, err error) {

	d, err := m.getDevice(iid)
	if err != nil {
		return
	}

	switch v := d.(type) {
	case OTADevice:
		return v.OTA(firmwareURL)
	default:
		logrus.Warnf("%s cant't OTA", iid)
		return
	}
}

func (m *Manager) Connect(iid string, params map[string]interface{}) (err error) {

	d, err := m.GetDevice(iid)
	if err != nil {
		return
	}

	if ad, authRequired := d.Device.(AuthDevice); authRequired && !ad.IsAuth() {
		if err = ad.Auth(params); err != nil {
			return
		}
	}

	if d.connected.Swap(true) {
		return
	}
	if err = d.Connect(); err != nil {
		return
	}

	d.df = definer.NewThingModelDefiner(iid, m.notifyAttr, m.notifyThingModelChange)
	d.Define(d.df)
	if d.df != nil {
		d.df.SetNotifyFunc()
		tm := d.df.ThingModel()

		// 记录所有设备（包括子设备）对应的 device
		for _, ins := range tm.Instances {
			m.devices.Store(ins.IID, d)
		}
	}

	return nil
}

func (m *Manager) Disconnect(iid string, params map[string]interface{}) (err error) {

	d, err := m.getDevice(iid)
	if err != nil {
		return
	}

	if ad, authRequired := d.(AuthDevice); authRequired && ad.IsAuth() {
		return ad.RemoveAuthorization(params)
	}
	return d.Disconnect(iid)
}

func (m *Manager) HealthCheck(iid string) bool {

	d, err := m.getDevice(iid)
	if err != nil {
		logrus.Warnf("device %s not found", iid)
		return false
	}

	online := d.Online(iid)
	logrus.Debugf("%s HealthCheck,online: %v", iid, online)
	return online
}

func (m *Manager) Devices() (ds []Device, err error) {
	return m.discoveredDevices, nil
}

func (m *Manager) notifyEvent(event Event) error {
	if m.eventChannel == nil {
		logrus.Warn("eventChannel not set")
		return nil
	}
	select {
	case m.eventChannel <- event:
	default:
	}

	logrus.Debugf("notifyEvent: %s:%s", event.Type, string(event.Data))
	return nil

}
func (m *Manager) notifyAttr(attrEvent definer.AttributeEvent) (err error) {
	data, _ := json.Marshal(attrEvent)
	ev := Event{
		Type: "attr_change",
		Data: data,
	}

	return m.notifyEvent(ev)
}

func (m *Manager) notifyThingModelChange(iid string, tme definer.ThingModelEvent) (err error) {

	d, loaded := m.devices.Load(iid)
	if !loaded {
		return nil
	}

	tme.ThingModel.OTASupport, err = m.IsOTASupport(iid)
	if err != nil {
		return
	}
	data, _ := json.Marshal(tme)
	ev := Event{
		Type: "thing_model_change",
		Data: data,
	}
	d.(*device).df.SetNotifyFunc()
	for _, ins := range tme.ThingModel.Instances {
		m.devices.Store(ins.IID, d)
	}
	return m.notifyEvent(ev)
}

func (m *Manager) SetAttributes(as []SetAttribute) error {

	for _, a := range as {
		df, err := m.getDefiner(a.IID)
		if err != nil {
			return err
		}
		if err = df.SetAttribute(a.IID, a.AID, a.Val); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) GetThingModel(iid string) (tm thingmodel.ThingModel, err error) {
	df, err := m.getDefiner(iid)
	if err != nil {
		return
	}
	return df.ThingModel(), nil
}

func (m *Manager) getDefiner(iid string) (df *definer.Definer, err error) {

	v, ok := m.devices.Load(iid)
	if !ok {
		return nil, errors.New("device not exist")
	}
	d := v.(*device)
	if d.df == nil {
		err = ThingModelNotFoundErr
		return
	}
	return d.df, nil
}

func (m *Manager) getDevice(iid string) (Device, error) {

	v, ok := m.devices.Load(iid)
	if !ok {
		return nil, errors.New("device not exist")
	}
	return v.(*device).Device, nil
}

func (m *Manager) GetDevice(iid string) (*device, error) {

	v, ok := m.devices.Load(iid)
	if !ok {
		return nil, errors.New("device not exist")
	}
	return v.(*device), nil
}
