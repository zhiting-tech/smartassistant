package device

import (
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-pinyin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
)

type ModelDevice struct {
	Name         string `json:"name"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	Logo         string `json:"logo"`         // logo地址
	Provisioning string `json:"provisioning"` // 配置页地址
	PluginID     string `json:"plugin_id"`

	Protocol string `json:"protocol"` // 连接云端的协议类型，tcp/mqtt
}

type Type struct {
	Name string            `json:"name"`
	Type plugin.DeviceType `json:"type"`
}

type Types []Type

type MajorResp struct {
	Types `json:"types"`
}

var majorTypes = map[plugin.DeviceType]string{
	plugin.TypeLight:          "照明",
	plugin.TypeSwitch:         "开关",
	plugin.TypeOutlet:         "插座",
	plugin.TypeRoutingGateway: "路由网关",
	plugin.TypeSecurity:       "安防",
	plugin.TypeSensor:         "传感器",
	plugin.TypeLifeElectric:   "生活电器",
}

// MajorTypeList 获取主分类
func MajorTypeList(c *gin.Context) {
	var (
		err  error
		resp MajorResp
	)
	resp.Types = make([]Type, 0)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	deviceConfigs := plugin.GetGlobalClient().DeviceConfigs()
	m := make(map[plugin.DeviceType]string, 0)
	for _, d := range deviceConfigs {
		if d.Provisioning == "" { // 没有配置置网页则忽略
			continue
		}

		pType := minorTypes[d.Type].ParentType
		if pType == "" {
			continue
		}

		m[pType] = majorTypes[pType]
	}

	for k, v := range m {
		resp.Types = append(resp.Types, Type{v, k})
	}

	sort.Sort(resp.Types) // 按拼音首字母A-Z排序
}

// 获取首字母Ascii码
func getInitialAscii(s string) int {
	ascii := []rune(s)
	return int(ascii[0])
}

// getInitialPinYin 获取拼音首字母
func getInitialPinyin(s string) string {
	py := pinyin.NewArgs()
	py.Style = pinyin.FirstLetter

	p := pinyin.Pinyin(s, py) // 获取拼音

	return p[0][0]
}

func (t Types) Len() int {
	return len(t)
}

func (t Types) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t Types) Less(i, j int) bool {
	iPinyin := getInitialPinyin(t[i].Name)
	iAscii := getInitialAscii(iPinyin)

	jPinyin := getInitialPinyin(t[j].Name)
	jAsciiI := getInitialAscii(jPinyin)

	return iAscii < jAsciiI
}
