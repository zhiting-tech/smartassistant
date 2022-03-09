package plugin

import (
	errors2 "errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/go-playground/validator/v10"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	version2 "github.com/zhiting-tech/smartassistant/modules/utils/version"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

type DeviceType string

const (
	TypeLight          DeviceType = "light"           // 照明
	TypeSwitch         DeviceType = "switch"          // 开关
	TypeOutlet         DeviceType = "outlet"          // 插座
	TypeRoutingGateway DeviceType = "routing_gateway" // 路由网关
	TypeSecurity       DeviceType = "security"        // 安防
	TypeSensor         DeviceType = "sensor"          // 传感器
	TypeLifeElectric   DeviceType = "life_electric"   // 生活电器

	TypeLamp             DeviceType = "lamp"               // 台灯
	TypeCeilingLamp      DeviceType = "ceiling_lamp"       // 吸顶灯
	TypeBulb             DeviceType = "bulb"               // 灯泡
	TypeLightStrip       DeviceType = "light_strip"        // 灯带
	TypePendantLight     DeviceType = "pendant_light"      // 吊灯
	TypeBedSideLamp      DeviceType = "bedside_lamp"       // 夜灯
	TypeNightLight       DeviceType = "night_light"        // 床头灯
	TypeFanLamp          DeviceType = "fan_lamp"           // 风扇灯
	TypeDownLight        DeviceType = "down_light"         // 简射灯
	TypeMagneticRailLamp DeviceType = "magnetic_rail_lamp" // 磁吸轨道灯

	TypeOneKeySwitch    DeviceType = "one_key_switch"   // 单键开关
	TypeTwoKeySwitch    DeviceType = "two_key_switch"   // 双键开关
	TypeThreeKeySwitch  DeviceType = "three_key_switch" // 三键开关
	TypeDWirelessSwitch DeviceType = "wireless_switch"  // 无线开关

	TypeConverter  DeviceType = "converter"   // 转换器
	TypeWallPlug   DeviceType = "wall_plug"   // 入墙插座
	TypePowerStrip DeviceType = "power_strip" // 排座

	TypeRouter       DeviceType = "router"        // 路由器
	TypeWifiRepeater DeviceType = "wifi_repeater" // wifi信号放大器
	TypeGateway      DeviceType = "gateway"       // 网关

	TypeCamera           DeviceType = "camera"            // 摄像头
	TypePeepholeDoorbell DeviceType = "peephole_doorbell" // 猫眼门铃
	TypeDoorLock         DeviceType = "door_lock"         // 门锁

	TypeCurtain DeviceType = "curtain" // 窗帘电机

	TypeTemperatureAndHumiditySensor DeviceType = "temperature_humidity_sensor" // 温湿度传感器
	TypeHumanSensors                 DeviceType = "human_sensor"                // 人体传感器
	TypeSmokeSensor                  DeviceType = "smoke_sensor"                // 烟雾传感器
	TypeGasSensor                    DeviceType = "gas_sensor"                  // 燃气传感器
	TypeWindowDoorSensor             DeviceType = "window_door_sensor"          // 门窗传感器
	TypeWaterLeakSensor              DeviceType = "water_leak_sensor"           // 水浸传感器
	TypeIlluminanceSensor            DeviceType = "illuminance_sensor"          // 光照度传感器
	TypeDynamicAndStaticSensor       DeviceType = "dynamic_static_sensor"       // 动静传感器
)

const (
	cpuShare = 512 // CPU资源紧张时插件CPU资源权重设为为512，保证不与SA服务抢资源。SA为默认1024

	// 两个参数绑定一起，标识只允许使用0.25个cpu资源
	cpuPeriod = 100000
	cpuQuota  = 2500

	networkModeHost = "host"
)

var (
	alwaysRestart = container.RestartPolicy{
		Name: "always",
	}
)

// Config 插件配置
type Config struct {
	Name           string         `yaml:"name" json:"name" validate:"required"`                       // 插件名称
	Version        string         `yaml:"version" json:"version" validate:"required"`                 // 版本
	Info           string         `yaml:"info" json:"info"`                                           // 介绍
	SupportDevices []DeviceConfig `yaml:"support_devices" json:"support_devices" validate:"required"` // 支持的设备
}

// ID 根据配置生成插件ID
func (p Config) ID() string {
	return p.Name
}
func (p Config) Validate() error {
	defaultValidator := validator.New()
	defaultValidator.SetTagName("validate")
	return defaultValidator.Struct(p)
}

// Plugin 插件详情
type Plugin struct {
	Config `yaml:",inline"`
	ID     string `json:"id" yaml:"id"`
	Brand  string `json:"brand" yaml:"brand"`
	Image  string `json:"image" yaml:"image"`
	Source string `json:"source" yaml:"source"` // 插件来源
	AreaID uint64 `json:"area_id" yaml:"area_id"`
}

func NewFromEntity(p entity.PluginInfo) Plugin {
	return Plugin{
		Config: Config{
			Name:    p.PluginID,
			Version: p.Version,
			Info:    p.Info,
		},
		ID:     p.PluginID,
		Image:  p.Image,
		AreaID: p.AreaID,
		Source: p.Source,
	}
}

// IsDevelopment 是否开发者上传的插件
func (p Plugin) IsDevelopment() bool {
	return p.Source == entity.SourceTypeDevelopment
}

func (p Plugin) IsAdded() bool {
	// return docker.GetClient().IsImageAdd(p.Image.RefStr())
	return entity.IsPluginAdd(p.ID, p.AreaID)
}
func (p Plugin) IsNewest() bool {
	if p.Source == entity.SourceTypeDevelopment {
		return true
	}
	return false // 方便开发更新插件

	pluginInfo, err := entity.GetPlugin(p.ID, p.AreaID)
	if err != nil {
		logger.Errorf("get plugin info fail: %v\n", err)
		return true
	}
	greater, err := version2.Greater(p.Version, pluginInfo.Version)
	if err != nil {
		logger.Errorf("compare plugin version fail: %v\n", err)
		return true
	}
	return greater
}

func (p Plugin) IsRunning() bool {
	isRunning, _ := docker.GetClient().ContainerIsRunningByImage(p.Image)
	return isRunning
}

// Up 启动插件
func (p Plugin) Up() (err error) {
	logger.Info("up plugin:", p.Name)
	_, err = RunPlugin(p)
	if err != nil && strings.Contains(err.Error(), "already in use") {
		return nil
	}
	return err
}
func (p Plugin) UpdateOrInstall() (err error) {
	if !p.IsDevelopment() {
		if err = docker.GetClient().Pull(p.Image); err != nil {
			return errors.Wrap(err, status.PluginPullFail)
		}
	}
	if p.IsAdded() {
		return p.Update()
	}
	return p.Install()
}

// Install 安装并且启动插件
func (p Plugin) Install() (err error) {

	// TODO 镜像没build或者build失败则不能安装

	if err = p.Up(); err != nil {
		return errors.Wrap(err, status.PluginUpFail)
	}

	var pi = entity.PluginInfo{
		AreaID:   p.AreaID,
		PluginID: p.ID,
		Image:    p.Image,
		Info:     p.Info,
		Status:   entity.StatusInstallSuccess,
		Version:  p.Version,
		Source:   p.Source,
		Brand:    p.Brand,
	}
	if err = entity.SavePluginInfo(pi); err != nil {
		logger.Errorf("UpdatePluginStatus err: %s", err.Error())
		return
	}
	return
}

// Update 更新插件
func (p Plugin) Update() (err error) {
	if p.Source == entity.SourceTypeDevelopment {
		return errors2.New("plugin in development can't update")
	}
	logger.Info("update plugin:", p.ID)

	if err = p.StopAndRemovePluginImage(); err != nil {
		logger.Error(err.Error())
	}
	return p.Install()
}

// StopAndRemovePluginImage 停止插件容器并删除插件镜像
func (p Plugin) StopAndRemovePluginImage() (err error) {

	// 查询数据库获取当前插件的镜像
	plgInfo, err := entity.GetPlugin(p.ID, p.AreaID)
	if err != nil {
		return
	}
	if err = docker.GetClient().StopContainer(plgInfo.Image); err != nil {
		return
	}
	if err = docker.GetClient().RemoveContainer(plgInfo.Image); err != nil {
		return
	}

	if plgInfo.Image == p.Image {
		return
	}
	if err = docker.GetClient().ImageRemove(plgInfo.Image); err != nil {
		return
	}
	return
}

// Remove 删除插件
func (p Plugin) Remove() (err error) {
	logger.Info("removing plugin", p.ID)

	// 先移除配对再暂停和移除容器
	devices, err := entity.GetDevicesByPluginID(p.ID)
	if err != nil {
		return
	}

	for _, d := range devices {
		if err = DisconnectDevice(d.Identity, d.PluginID, nil); err != nil {
			logger.Error("disconnect err:", err)
		}
	}

	if err = p.StopAndRemovePluginImage(); err != nil {
		logger.Error(err.Error())
	}

	if err = entity.DelDevicesByPlgID(p.ID); err != nil {
		return
	}

	if err = entity.DelPlugin(p.ID, p.AreaID); err != nil {
		return
	}
	return
}

type Attribute struct {
	server.Attribute
	CanControl bool `json:"can_control"`
}

type Instance struct {
	Type       string      `json:"type"`
	InstanceId int         `json:"instance_id"`
	Attributes []Attribute `json:"attributes"`
}

type DeviceInstances struct {
	Identity  string     `json:"identity"`
	Instances []Instance `json:"instances"`

	// OTASupport 是否支持OTA
	OTASupport bool `json:"ota_support"`
}

type DeviceInfo struct {
	Identity     string
	Model        string
	Manufacturer string
	Version      string
}

// GetInfo 获取设备基础信息
func (das DeviceInstances) GetInfo() (d DeviceInfo, err error) {

	for _, ins := range das.Instances {
		if ins.Type != "info" {
			continue
		}
		for _, attr := range ins.Attributes {
			switch attr.Attribute.Attribute {
			case "model":
				d.Model = attr.Val.(string)
			case "manufacturer":
				d.Manufacturer = attr.Val.(string)
			case "identity":
				d.Identity = attr.Val.(string)
			case "version":
				d.Version = attr.Val.(string)
			}
		}
		return
	}
	err = errors2.New("info instance not found")
	return
}

func GetInfoFromDeviceAttrs(pluginID string, das DeviceInstances) (d entity.Device, err error) {
	d.Identity = das.Identity
	d.PluginID = pluginID
	for _, ins := range das.Instances {
		if ins.Type == "info" {
			for _, attr := range ins.Attributes {
				switch attr.Attribute.Attribute {
				case "model":
					d.Model = attr.Val.(string)
					d.Name = attr.Val.(string)
				case "manufacturer":
					d.Manufacturer = attr.Val.(string)
				}
			}
			return
		}
	}
	err = errors2.New("no instance info found")
	return
}

type OnDeviceStateChange func(d entity.Device, attr entity.Attribute) error

func DefaultOnDeviceStateChange(d entity.Device, attr entity.Attribute) error {
	return errors2.New("OnDeviceStateChange not implement")
}

type DiscoverResponse struct {
	Name         string `json:"name"`
	Identity     string `json:"identity"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	PluginID     string `json:"plugin_id"`
	LogoURL      string `json:"logo_url"`
	AuthRequired bool   `json:"auth_required"`
}

// RunPlugin 运行插件
func RunPlugin(plg Plugin) (containerID string, err error) {
	conf := container.Config{
		Image: plg.Image,
		Env:   []string{fmt.Sprintf("PLUGIN_DOMAIN=%s", plg.ID)},
	}
	// 映射插件目录到宿主机上
	source := filepath.Join(config.GetConf().SmartAssistant.RuntimePath,
		"data", "plugin", plg.Brand, plg.Name)
	if err = os.MkdirAll(source, os.ModePerm); err != nil {
		return
	}
	target := "/app/data/"
	logger.Debugf("mount %s to %s", source, target)

	// 需要使用宿主机能识别的路径来挂载，TODO 当前实现可能导致混乱,再后面优化
	hostSource := filepath.Join(config.GetConf().SmartAssistant.HostRuntimePath,
		"data", "plugin", plg.Brand, plg.Name)
	hostConf := container.HostConfig{
		NetworkMode:   networkModeHost,
		RestartPolicy: alwaysRestart,
		Mounts: []mount.Mount{
			{Type: mount.TypeBind, Source: hostSource, Target: target},
		},
		Resources: container.Resources{
			CPUShares: cpuShare,
			CPUPeriod: cpuPeriod,
			CPUQuota:  cpuQuota,
		},
	}
	if config.GetConf().SmartAssistant.FluentdAddress != "" {
		// 设置容器的logging, driver
		hostConf.LogConfig = container.LogConfig{
			Type: "fluentd",
			Config: map[string]string{
				"fluentd-address": config.GetConf().SmartAssistant.FluentdAddress,
				"tag":             fmt.Sprintf("smartassistant.plugin.%s", plg.Image),
			},
		}
	}
	return docker.GetClient().ContainerRun(plg.Image, conf, hostConf)
}
