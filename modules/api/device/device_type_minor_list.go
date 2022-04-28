package device

import (
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type MinorResp struct {
	Types MinorTypes `json:"types"`
}

type MinorTypes []MinorType

type MinorType struct {
	Type
	Devices []ModelDevice `json:"devices"`
}

type MinorReq struct {
	Type plugin.DeviceType `form:"type"`
}

type minorType struct {
	Name       string
	ParentType plugin.DeviceType
}

var minorTypes = map[plugin.DeviceType]minorType{
	plugin.TypeLamp:             {"台灯", plugin.TypeLight},
	plugin.TypeCeilingLamp:      {"吸顶灯", plugin.TypeLight},
	plugin.TypeBulb:             {"灯泡", plugin.TypeLight},
	plugin.TypeLightStrip:       {"灯带", plugin.TypeLight},
	plugin.TypePendantLight:     {"吊灯", plugin.TypeLight},
	plugin.TypeBedSideLamp:      {"夜灯", plugin.TypeLight},
	plugin.TypeNightLight:       {"床头灯", plugin.TypeLight},
	plugin.TypeFanLamp:          {"风扇灯", plugin.TypeLight},
	plugin.TypeDownLight:        {"简射灯", plugin.TypeLight},
	plugin.TypeMagneticRailLamp: {"磁吸轨道灯", plugin.TypeLight},

	plugin.TypeOneKeySwitch:   {"单键开关", plugin.TypeSwitch},
	plugin.TypeTwoKeySwitch:   {"双键开关", plugin.TypeSwitch},
	plugin.TypeThreeKeySwitch: {"三键开关", plugin.TypeSwitch},
	plugin.TypeWirelessSwitch: {"无线开关", plugin.TypeSwitch},

	plugin.TypeConverter:  {"转换器", plugin.TypeOutlet},
	plugin.TypeWallPlug:   {"入墙插座", plugin.TypeOutlet},
	plugin.TypePowerStrip: {"排座", plugin.TypeOutlet},

	plugin.TypeRouter:       {"路由器", plugin.TypeRoutingGateway},
	plugin.TypeWifiRepeater: {"Wi-Fi信号放大器", plugin.TypeRoutingGateway},
	plugin.TypeGateway:      {"网关", plugin.TypeRoutingGateway},

	plugin.TypeCamera:           {"摄像头", plugin.TypeSecurity},
	plugin.TypePeepholeDoorbell: {"猫眼门铃", plugin.TypeSecurity},
	plugin.TypeDoorLock:         {"门锁", plugin.TypeSecurity},

	plugin.TypeCurtain: {"窗帘电机", plugin.TypeLifeElectric},

	plugin.TypeTemperatureAndHumiditySensor: {"温湿度传感器", plugin.TypeSensor},
	plugin.TypeHumanSensors:                 {"人体传感器", plugin.TypeSensor},
	plugin.TypeSmokeSensor:                  {"烟雾传感器", plugin.TypeSensor},
	plugin.TypeGasSensor:                    {"燃气传感器", plugin.TypeSensor},
	plugin.TypeWindowDoorSensor:             {"门窗传感器", plugin.TypeSensor},
	plugin.TypeWaterLeakSensor:              {"水浸传感器", plugin.TypeSensor},
	plugin.TypeIlluminanceSensor:            {"光照度传感器", plugin.TypeSensor},
	plugin.TypeDynamicAndStaticSensor:       {"动静传感器", plugin.TypeSensor},
}

// MinorTypeList 根据主分类获取次级分类和设备类型
func MinorTypeList(c *gin.Context) {
	var (
		err  error
		resp MinorResp
		req  MinorReq
	)
	resp.Types = make(MinorTypes, 0)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindQuery(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if _, ok := majorTypes[req.Type]; !ok {
		err = errors.Wrap(err, status.DeviceTypeNotExist)
		return
	}

	var token string
	u := session.Get(c)
	if u != nil {
		token = u.Token
	}
	deviceConfigs := plugin.GetGlobalClient().DeviceConfigs()
	m := make(map[plugin.DeviceType][]ModelDevice)
	for _, d := range deviceConfigs {
		if d.Provisioning == "" { // 没有配置置网页则忽略
			continue
		}
		// 拼接token和插件id辅助插件实现websocket请求
		provisioning, err := d.WrapProvisioning(token, d.PluginID)
		if err != nil {
			logger.Error(err)
			continue
		}
		md := ModelDevice{
			Name:  d.Name,
			Model: d.Model,
			Logo: plugin.PluginTargetURL(c.Request, d.PluginID,
				d.Model, d.Logo), // 根据配置拼接插件中的图片地址
			Provisioning: provisioning,
			PluginID:     d.PluginID,
			Protocol:     d.Protocol,
		}
		if req.Type == minorTypes[d.Type].ParentType || req.Type == d.Type {
			m[d.Type] = append(m[d.Type], md)
		}
	}

	for k, v := range m {
		name := minorTypes[k].Name
		if name == "" {
			name = "其他"
		}
		resp.Types = append(resp.Types, MinorType{Type{name, k}, v})
	}

	sort.Sort(resp.Types) // 按拼音首字母A-Z排序
}

func (t MinorTypes) Len() int {
	return len(t)
}

func (t MinorTypes) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t MinorTypes) Less(i, j int) bool {
	iPinyin := getInitialPinyin(t[i].Name)
	iAscii := getInitialAscii(iPinyin)

	jPinyin := getInitialPinyin(t[j].Name)
	jAsciiI := getInitialAscii(jPinyin)

	return iAscii < jAsciiI
}
