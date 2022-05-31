package sdk

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

var ThingModelNotFoundErr = errors.New("thing model not found")

func newDevice(d Device) *device {
	dd := device{
		Device: d,
	}
	return &dd
}

type device struct {
	Device
	df *definer.Definer

	connected bool
	mutex     sync.Mutex
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
func (m *Manager) InitOrUpdateDevice(d Device) error {
	if d == nil {
		return errors.New("device is nil")
	}

	v, loaded := m.devices.LoadOrStore(d.Info().IID, newDevice(d))
	if loaded {
		logrus.Debugf("device %s already exist", d.Info().IID)
		oldDevice := v.(*device)
		if oldDevice.Address() != d.Address() {
			logrus.Debugf("device %v change address: %s -> %s:",
				d.Info().IID, oldDevice.Address(), d.Address())
			m.devices.Store(d.Info().IID, newDevice(d))
			oldDevice.Disconnect(d.Info().IID)
			return nil
		}

		return nil
	}
	logrus.Info("add device:", d.Info())

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

	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.connected {
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

	// 设置设备已连接，数据库标记为已添加
	d.connected = true

	return nil
}

func (m *Manager) Disconnect(iid string, params map[string]interface{}) (err error) {

	defer m.devices.Delete(iid)
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
	go func() {
		// 还没有连接则主动连接，设备需要自行处理认证信息持久化
		if err = m.Connect(iid, nil); err != nil {
			return
		}
	}()

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

	d, err := m.GetDevice(iid)
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
