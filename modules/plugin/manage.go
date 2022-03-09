package plugin

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// Manager 与SC服务交互获取插件信息
type Manager interface {
	// LoadPlugins 加载并返回所有插件
	LoadPlugins() (map[string]*Plugin, error)
	// GetPlugin 加载并返回插件
	GetPlugin(id string) (*Plugin, error)
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
}

// Client 与插件服务交互的客户端
type Client interface {
	DevicesDiscover(ctx context.Context) <-chan DiscoverResponse
	GetAttributes(device entity.Device) (DeviceInstances, error)
	SetAttributes(device entity.Device, data json.RawMessage) (result []byte, err error)
	IsOnline(entity.Device) bool

	OTA(d entity.Device, firmwareURL string) error

	// Connect 连接设备
	Connect(identity, pluginID string, authParams map[string]string) (DeviceInstances, error)
	// Disconnect 与设备断开连接
	Disconnect(identity, pluginID string, authParams map[string]string) error

	// DeviceConfig 设备的配置
	DeviceConfig(entity.Device) DeviceConfig
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
		globalClient = NewClient(DefaultOnDeviceStateChange)
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
	plugins, err := m.LoadPlugins()
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
