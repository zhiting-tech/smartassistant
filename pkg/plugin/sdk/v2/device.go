package sdk

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

type DeviceInfo struct {
	IID          string
	Model        string
	Manufacturer string
	Type         string
	AuthRequired bool
}

// Device 设备
type Device interface {

	// Address 设备地址，用来判断已添加设备是否变更了ip和端口，是的话回断掉旧的，连接新的
	Address() string

	// Connect 连接设备
	Connect() error

	// Disconnect 断开与设备的连接
	Disconnect(iid string) error

	// Define 使用 *definer.Definer 来生成物模型
	Define(definer2 *definer.Definer)

	// Info 设备信息，通常在局域网发现阶段可以获取
	Info() DeviceInfo

	// Online 设备或子设备是否在线
	Online(iid string) bool
}

// AuthDevice 需要认证的设备
type AuthDevice interface {
	Device
	// AuthParams 返回认证/配对参数的定义：类型、默认值、名字等
	AuthParams() []thingmodel.AuthParam
	// IsAuth 返回设备是否成功认证/配对
	IsAuth() bool
	// Auth 认证/配对
	Auth(params map[string]interface{}) error
	// RemoveAuthorization 取消认证/配对
	RemoveAuthorization(params map[string]interface{}) error
}

type OTAProgressState int // OTA进度

const (
	OTAUpgradeFail  OTAProgressState = -1 // 更新失败
	OTADownloadFail OTAProgressState = -2 // 下载失败
	OTAValidateFail OTAProgressState = -3 // 校验失败
	OTABurnFail     OTAProgressState = -4 // 烧写失败
)

func OTAProgress(i int) OTAProgressState {
	return OTAProgressState(i)
}

type OTAResp struct {
	Step OTAProgressState
}

type OTADevice interface {
	Device
	OTA(firmwareURL string) (chan OTAResp, error)
}

type AttrInfo struct {
	AID        int         `json:"aid"`
	Type       string      `json:"type"`
	ValType    string      `json:"val_type"`
	Min        interface{} `json:"min,omitempty"`
	Max        interface{} `json:"max,omitempty"`
	Permission uint        `json:"permission"`

	Val interface{} `json:"val"`
}
