package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

var (
	DeviceNotExist = errors.New("device not exist")
)

const (
	ThingModelChangeEvent = "thing_model_change"
	AttrChangeEvent       = "attr_change"
)

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
		addrChange := oldDevice.Address() != d.Address()
		if addrChange {
			logrus.Debugf("device %s change address: %s -> %s:",
				d.Info().IID, oldDevice.Address(), d.Address())
		}
		disconnected := oldDevice.connected && !oldDevice.Online(d.Info().IID)
		if disconnected {
			logrus.Debugf("device %s disconnected, reconnecting...", d.Info().IID)
		}
		if addrChange || disconnected {
			m.devices.Store(d.Info().IID, newDevice(d))
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

func (m *Manager) Connect(d *device, params map[string]interface{}) (err error) {

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

	logrus.Debugf("device %s connecting...", d.Info().IID)
	if err = d.Connect(); err != nil {
		logrus.Errorf("connect err: %s", err)
		return
	}

	logrus.Debugf("device %s connected, define device's thing model", d.Info().IID)
	d.df = definer.NewThingModelDefiner(d.Info().IID, m.notifyAttr, m.notifyThingModelChange)
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

	logrus.Debugf("device %s connect and define done", d.Info().IID)
	return nil
}

func (m *Manager) Disconnect(iid string, params map[string]interface{}) (err error) {

	defer m.devices.Delete(iid)
	d, err := m.getDevice(iid)
	if err != nil {
		return
	}

	if ad, authRequired := d.(AuthDevice); authRequired && ad.IsAuth() {
		if ad.Info().IID == iid {
			return ad.RemoveAuthorization(params)
		}
	}
	return d.Disconnect(iid)
}

func (m *Manager) HealthCheck(iid string) bool {

	d, err := m.GetDevice(iid)
	if err != nil {
		logrus.Warnf("device %s not found", iid)
		return false
	}
	go func() {
		// 还没有连接则主动连接，设备需要自行处理认证信息持久化
		if err = m.Connect(d, nil); err != nil {
			return
		}
	}()

	// 需要授权且未授权则更新物模型
	if ad, ok := d.Device.(AuthDevice); ok && !ad.IsAuth() {

		go func() {
			tme := definer.ThingModelEvent{
				ThingModel: thingmodel.ThingModel{},
				IID:        iid,
			}
			m.notifyThingModelChange(iid, tme)
		}()
		return false
	}

	online := d.Device.Online(iid)
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
		Type: AttrChangeEvent,
		Data: data,
	}

	return m.notifyEvent(ev)
}

// notifyThingModelChange iid是桥接设备iid，tme.IID是实际更新的设备iid
func (m *Manager) notifyThingModelChange(iid string, tme definer.ThingModelEvent) (err error) {

	tme.ThingModel.OTASupport, err = m.IsOTASupport(iid)
	d, err := m.GetDevice(iid)
	if err != nil {
		return
	}
	var ad AuthDevice
	ad, tme.ThingModel.AuthRequired = d.Device.(AuthDevice)
	if tme.ThingModel.AuthRequired {
		tme.ThingModel.IsAuth = ad.IsAuth()
		tme.ThingModel.AuthParams = ad.AuthParams()
	}
	if err != nil {
		return
	}
	data, _ := json.Marshal(tme)
	ev := Event{
		Type: ThingModelChangeEvent,
		Data: data,
	}
	// 物模型变更需要重新给新增加的instance设置通知函数
	if d.df != nil {
		d.df.SetNotifyFunc()
	}
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
	if err != nil {
		return
	}
	if d.df == nil {
		err = fmt.Errorf("%s definer is nil", iid)
		return
	}
	return d.df, nil
}

func (m *Manager) getDevice(iid string) (Device, error) {

	d, err := m.GetDevice(iid)
	if err != nil {
		return nil, err
	}
	return d.Device, nil
}

func (m *Manager) GetDevice(iid string) (*device, error) {

	v, ok := m.devices.Load(iid)
	if !ok {
		return nil, DeviceNotExist
	}
	if d, ok := v.(*device); ok {
		return d, nil
	}
	return nil, fmt.Errorf("%s: is not *device", iid)
}
