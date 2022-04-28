package plugin

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"

	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// Manager 与SC服务交互获取插件信息
type Manager interface {
	// LoadPluginsWithContext 加载并返回所有插件
	LoadPluginsWithContext(ctx context.Context) (map[string]*Plugin, error)
	// GetPluginWithContext 加载并返回插件
	GetPluginWithContext(ctx context.Context, id string) (*Plugin, error)
}

// DeviceConfig 插件设备配置
type DeviceConfig struct {
	Model string     `json:"model" yaml:"model"`
	Name  string     `json:"name" yaml:"name"`
	Type  DeviceType `json:"type" yaml:"type"` // 设备类型

	Logo         string `json:"logo" yaml:"logo"`                 // 设备logo相对路径
	Control      string `json:"control" yaml:"control"`           // 设备控制页面相对路径
	Provisioning string `json:"provisioning" yaml:"provisioning"` // 设备置网页面相对路径
	PluginID     string `json:"-"`

	// SupportGateways 支持网关列表，仅子设备时有值
	SupportGateways []Gateway `json:"support_gateways,omitempty"`

	// Protocol 设备连接云端的协议类型，TCP/MQTT，当前仅美汇智居设备支持配置
	Protocol string `json:"protocol,omitempty"`
}

// WrapProvisioning 包装设备添加页，添加参数。TODO 当前实现由于插件配置的地址中含有#符号，需要特殊处理
func (dc DeviceConfig) WrapProvisioning(token, pluginID string) (provisioning string, err error) {

	u, err := url.Parse(dc.Provisioning)
	if err != nil {
		return
	}
	if u.Fragment != "" {
		var q url.Values
		q, err = url.ParseQuery(u.Fragment)
		if err != nil {
			return
		}
		q.Add("token", token)
		q.Add("plugin_id", pluginID)
		u.Fragment, err = url.QueryUnescape(q.Encode())
		if err != nil {
			return
		}
	} else {
		q := u.Query()
		q.Add("token", token)
		q.Add("plugin_id", pluginID)
		u.RawQuery, err = url.QueryUnescape(q.Encode())
		if err != nil {
			return
		}
	}
	return u.String(), nil
}
func (dc DeviceConfig) IsGatewaySupport(gatewayModel string) bool {
	for _, gw := range dc.SupportGateways {
		if gw.Model == gatewayModel {
			return true
		}
	}
	return false
}

type Gateway struct {
	Name  string `json:"name"`
	Model string `json:"model"`
}

// Identify 设备唯一标识
type Identify struct {
	PluginID string `json:"plugin_id"`
	IID      string `json:"iid"`
	AreaID   uint64 `json:"area_id"`
}

func (id Identify) ID() string {
	return fmt.Sprintf("%s_%d_%s", id.PluginID, id.AreaID, id.IID)
}

// Client 与插件服务交互的客户端
type Client interface {
	DevicesDiscover(ctx context.Context) <-chan DiscoverResponse
	GetAttributes(ctx context.Context, identify Identify) (thingmodel.ThingModel, error)
	SetAttributes(ctx context.Context, pluginID string, areaID uint64, setReq sdk.SetRequest) (result []byte, err error)
	IsOnline(identify Identify) bool

	OTA(ctx context.Context, identify Identify, firmwareURL string) error

	// Connect 连接设备
	Connect(ctx context.Context, identify Identify, authParams map[string]interface{}) (thingmodel.ThingModel, error)
	// Disconnect 与设备断开连接
	Disconnect(ctx context.Context, identify Identify, authParams map[string]interface{}) error

	// DeviceConfig 设备的配置
	DeviceConfig(pluginID, model string) DeviceConfig
	// DeviceConfigs 所有设备的配置
	DeviceConfigs() []DeviceConfig
}

type Info struct {
	Logo         string `json:"logo" yaml:"control"`              // 设备logo地址相对路径
	Control      string `json:"control" yaml:"control"`           // 设备控制页面相对路径
	Provisioning string `json:"provisioning" yaml:"provisioning"` // 设备置网页面相对路径
	Compress     string `json:"compress" yaml:"compress"`         // 压缩包地址
}

var (
	globalManager     Manager
	globalManagerOnce sync.Once

	globalClient     Client
	globalClientOnce sync.Once
)

func SetGlobalClient(c Client) {
	globalClientOnce.Do(func() {
		globalClient = c
	})
}

func GetGlobalClient() Client {
	globalClientOnce.Do(func() {
		globalClient = NewClient()
	})
	return globalClient
}

func GetGlobalManager() Manager {
	globalManagerOnce.Do(func() {
		globalManager = NewManager()
		loadAndUpPlugins(globalManager)
	})
	return globalManager
}

func SetGlobalManager(m Manager) {
	globalManagerOnce.Do(func() {
		globalManager = m
		loadAndUpPlugins(globalManager)
	})
}

// loadAndUpPlugins 加载插件并启动已安装的插件
func loadAndUpPlugins(m Manager) {

	logger.Info("starting plugin globalManager")
	// 加载插件列表
	plugins, err := m.LoadPluginsWithContext(context.TODO())
	if err != nil {
		return
	}
	// 扫描已安装的插件，并且启动，连接 state change...
	// 等待其他容器启动，判断如果插件没有运行，则启动
	time.Sleep(5 * time.Second)
	for _, plg := range plugins {
		if !plg.IsAdded() || plg.IsRunning() {
			continue
		}
		// 如果镜像没运行，则启动
		if upErr := plg.Up(); upErr != nil {
			logger.Error("plugin up error:", upErr)
		}
	}
}
