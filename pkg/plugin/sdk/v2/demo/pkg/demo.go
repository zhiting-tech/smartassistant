package pkg

import (
	"time"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

// DemoProtocolDevice 描述硬件的协议
type DemoProtocolDevice struct {
	id           string
	model        string
	manufacturer string

	gateway  *definer.BaseService
	light    *definer.BaseService
	switches map[string]*definer.BaseService
}

func (sd *DemoProtocolDevice) Connect() error {
	return nil
}

func (sd *DemoProtocolDevice) Disconnect(iid string) error {
	return nil
}

func (sd *DemoProtocolDevice) Define(def *definer.Definer) {
	sd.id = "id111"
	// 连接设备，根据协议描述物模型

	// gateway
	sd.gateway = def.Instance("id111").NewGateway()
	sd.gateway.Enable(thingmodel.OnOff, OnOff{}).SetVal("on")

	// sub device light
	sd.light = def.Instance("id222").NewLight()
	sd.light.Enable(thingmodel.OnOff, OnOff{}).SetVal("on")
	sd.light.Enable(thingmodel.Brightness, Brightness{}).SetRange(1, 100).SetVal(55)

	// sub device switch
	switchInstance := def.Instance("id333")
	ls := switchInstance.NewSwitch()
	ls.Enable(thingmodel.OnOff, NewAttribute("on_off", "left")).SetVal("on")
	ms := switchInstance.NewSwitch()
	ms.Enable(thingmodel.OnOff, NewAttribute("on_off", "middle")).SetVal("off")
	rs := switchInstance.NewSwitch()
	rs.Enable(thingmodel.OnOff, NewAttribute("on_off", "right")).SetVal("on")
	sd.switches["left"] = ls
	sd.switches["middle"] = ms
	sd.switches["right"] = rs
	go sd.Listen(def)
	return
}

func (sd DemoProtocolDevice) Info() sdk.DeviceInfo {
	return sdk.DeviceInfo{
		IID:          sd.id,
		Model:        sd.model,
		Manufacturer: sd.manufacturer,
	}
}

// Online 在线 TODO 子设备的在线状态判断
func (sd *DemoProtocolDevice) Online(iid string) bool {
	return true
}

func (sd *DemoProtocolDevice) Listen(def *definer.Definer) {

	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			_ = sd.light.Notify(thingmodel.OnOff, "on")
			_ = def.UpdateThingModel()
		}
	}
}

func (sd *DemoProtocolDevice) AuthParams() []sdk.AuthParam {
	return nil
}

func (sd *DemoProtocolDevice) IsAuth() bool {
	return false
}

func (sd *DemoProtocolDevice) Auth(params map[string]interface{}) error {
	return nil
}

func (sd *DemoProtocolDevice) RemoveAuthorization(params map[string]interface{}) error {
	return nil
}

func (sd *DemoProtocolDevice) Close() error {
	return nil
}

func NewDemo(id string) *DemoProtocolDevice {
	return &DemoProtocolDevice{
		id:           id,
		manufacturer: "demo",
		switches:     make(map[string]*definer.BaseService),
	}
}
