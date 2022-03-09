package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/event"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"

	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
)

var NotExistErr = errors.New("plugin not exist")

const (
	healthCheckInterval = time.Second * 10 // 存活检查间隔
	offlineTimeout      = time.Second * 15 // 超过这个时间没有响应在线判断为离线
)

type client struct {
	mu      sync.Mutex // clients 锁
	clients map[string]*pluginClient

	devicesCancel        sync.Map
	stateChangeCallbacks []OnDeviceStateChange
}

func (c *client) DeviceConfigs() (configs []DeviceConfig) {

	for _, cli := range c.clients {
		for _, d := range cli.PluginConf.SupportDevices {
			d.PluginID = cli.pluginID
			configs = append(configs, d)
		}
	}
	return
}

func (c *client) DeviceConfig(d entity.Device) (config DeviceConfig) {
	if d.Model == types.SaModel {
		return
	}

	cli, err := c.get(d.PluginID)
	if err != nil {
		return
	}

	for _, sd := range cli.PluginConf.SupportDevices {
		if d.Model != sd.Model {
			continue
		}
		sd.PluginID = cli.pluginID
		return sd
	}
	return
}

func (c *client) Disconnect(identity, pluginID string, authParams map[string]string) (err error) {
	req := proto.AuthReq{
		Identity: identity,
		Params:   authParams,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	cli, err := c.get(pluginID)
	if err != nil {
		return
	}
	_, err = cli.protoClient.Disconnect(ctx, &req)
	if err != nil {
		return
	}

	v, loaded := c.devicesCancel.LoadAndDelete(identity)
	if loaded {
		if cancel, ok := v.(context.CancelFunc); ok {
			cancel()
		}
	}
	return nil
}

func NewClient(callbacks ...OnDeviceStateChange) *client {
	return &client{
		clients:              make(map[string]*pluginClient),
		stateChangeCallbacks: callbacks,
	}
}

func (c *client) get(domain string) (*pluginClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cli, ok := c.clients[domain]
	if ok {
		return cli, nil
	}
	return nil, NotExistErr
}

func (c *client) Add(cli *pluginClient) {

	c.mu.Lock()
	c.clients[cli.pluginID] = cli
	c.mu.Unlock()
	go c.ListenStateChange(cli.pluginID)
}

func (c *client) Remove(pluginID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cli, ok := c.clients[pluginID]
	if ok {
		delete(c.clients, pluginID)
		go cli.Stop()
	}
	return nil
}

// DevicesDiscover 发现设备，并且通过 channel 返回给调用者
func (c *client) DevicesDiscover(ctx context.Context) <-chan DiscoverResponse {
	out := make(chan DiscoverResponse, 10)
	go func() {
		var wg sync.WaitGroup
		for _, cli := range c.clients {
			wg.Add(1)
			go func(cli *pluginClient) {
				defer wg.Done()
				logger.Debug("listening plugin Discovering...")
				cli.DeviceDiscover(ctx, out)
				logger.Debug("plugin listening done")
			}(cli)
		}
		wg.Wait()
		close(out)
	}()
	return out
}

func (c *client) ListenStateChange(pluginID string) {
	cli, err := c.get(pluginID)
	if err != nil {
		return
	}
	pdc, err := cli.protoClient.StateChange(cli.ctx, &proto.Empty{})
	if err != nil {
		logger.Error("state onDeviceStateChange error:", err)
		return
	}
	logger.Println("StateChange recv...")
	for {
		var resp *proto.State
		resp, err = pdc.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error(err)
			// TODO retry
			break
		}
		logger.Debugf("get state onDeviceStateChange resp: %s,%d,%s\n",
			resp.Identity, resp.InstanceId, string(resp.Attributes))

		go func() {
			defer func() {
				if r := recover(); r != nil {
					logrus.Error(r)
				}
			}()
			var d entity.Device

			// 查看是否存在对应的子设备, 如果存在，则本次推送，是子设备
			sonDeviceIdentity := fmt.Sprintf("childDevice-%s-%d", resp.Identity, resp.InstanceId)
			sonDevice, err := entity.GetPluginDevice(cli.areaID, cli.pluginID, sonDeviceIdentity)
			if err == nil {
				d = sonDevice
			} else {
				// 如果不存在对应的子设备，则代表推送的是本身的设备
				d, err = entity.GetPluginDevice(cli.areaID, cli.pluginID, resp.Identity)
				if err != nil {
					logger.Errorf("ListenStateChange error:%s", err.Error())
					return
				}
			}

			var attr server.Attribute
			_ = json.Unmarshal(resp.Attributes, &attr)
			a := entity.Attribute{
				Attribute:  attr,
				InstanceID: int(resp.InstanceId),
			}

			em := event.NewEventMessage(event.AttributeChange, d.AreaID)
			em.SetDeviceID(d.ID)
			em.SetAttr(a)
			event.GetServer().Notify(em)
		}()
	}
	logger.Println("StateChangeFromPlugin exit")
}

func (c *client) SetAttributes(d entity.Device, data json.RawMessage) (result []byte, err error) {
	req := proto.SetAttributesReq{
		Identity: d.Identity,
		Data:     data,
	}
	logger.Debug("set attributes: ", string(data))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	cli, err := c.get(d.PluginID)
	if err != nil {
		return
	}
	_, err = cli.protoClient.SetAttributes(ctx, &req)
	if err != nil {
		logger.Error(err)
		return
	}
	return
}

func (c *client) OTA(d entity.Device, firmwareURL string) (err error) {
	req := proto.OTAReq{
		Identity:    d.Identity,
		FirmwareUrl: firmwareURL,
	}
	logger.Debugf("ota: %s, firmware url: %s", d.Identity, firmwareURL)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	cli, err := c.get(d.PluginID)
	if err != nil {
		return
	}
	otaCli, err := cli.protoClient.OTA(ctx, &req)
	if err != nil {
		return
	}

	for {
		var resp *proto.OTAResp
		resp, err = otaCli.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return errors.New("ota err, eof")
			}
			return
		}
		logrus.Debug("ota response:", resp.Identity, resp.Step)
		if resp.Step >= 100 {
			return nil
		}
		if resp.Step < 0 {
			return fmt.Errorf("ota err, step: %d", resp.Step)
		}
	}
}

// Connect 连接设备
func (c *client) Connect(identity, pluginID string, authParams map[string]string) (das DeviceInstances, err error) {
	req := proto.AuthReq{
		Identity: identity,
		Params:   authParams,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5) // 配对认证过程较长可能长达数分钟
	defer cancel()

	cli, err := c.get(pluginID)
	if err != nil {
		return
	}
	resp, err := cli.protoClient.Connect(ctx, &req)
	if err != nil {
		return
	}
	logger.Debugf("connect resp: %#v\n", resp)
	das = parseAttrsResp(identity, resp)
	return
}

func parseAttrsResp(identity string, resp *proto.GetAttributesResp) DeviceInstances {

	var instances []Instance
	for _, instance := range resp.Instances {
		var attrs []Attribute
		_ = json.Unmarshal(instance.Attributes, &attrs)
		i := Instance{
			Type:       instance.Type,
			InstanceId: int(instance.InstanceId),
			Attributes: attrs,
		}
		instances = append(instances, i)
	}
	return DeviceInstances{
		Identity:   identity,
		Instances:  instances,
		OTASupport: resp.OtaSupport,
	}
}

func (c *client) GetAttributes(d entity.Device) (das DeviceInstances, err error) {
	req := proto.GetAttributesReq{Identity: d.Identity}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cli, err := c.get(d.PluginID)
	if err != nil {
		return
	}
	resp, err := cli.protoClient.GetAttributes(ctx, &req)
	if err != nil {
		return
	}
	logger.Debugf("GetAttributes resp: %#v\n", resp)
	das = parseAttrsResp(d.Identity, resp)
	return
}

func (c *client) IsOnline(d entity.Device) bool {
	cli, err := c.get(d.PluginID)
	if err != nil {
		logger.Warningf("get plugin %s error: %s", d.PluginID, err)
		return false
	}
	return cli.IsOnline(d.Identity)
}

func newClient(areaID uint64, plgID, key string, plgConf Plugin) (*pluginClient, error) {
	cli, err := clientv3.NewFromURL(etcdURL)
	if err != nil {
		return nil, err
	}
	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(fmt.Sprintf("etcd:///%s", key), grpc.WithInsecure(), grpc.WithResolvers(etcdResolver))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &pluginClient{
		areaID:      areaID,
		pluginID:    plgID,
		protoClient: proto.NewPluginClient(conn),
		ctx:         ctx,
		cancel:      cancel,
		PluginConf:  plgConf,
	}, nil
}

type pluginClient struct {
	areaID      uint64
	pluginID    string
	protoClient proto.PluginClient // 请求插件服务的grpc客户端
	cancel      context.CancelFunc
	ctx         context.Context
	PluginConf  Plugin

	healthCheckDevice    sync.Map // 记录有存活检查的设备，防止重复发起存活检查
	deviceLastOnlineTime sync.Map // 记录每个设备上次在线时间
}

func (pc *pluginClient) Stop() {
	if pc.cancel != nil {
		pc.cancel()
	}
}

func (pc *pluginClient) DeviceDiscover(ctx context.Context, out chan<- DiscoverResponse) {

	pdc, err := pc.protoClient.Discover(ctx, &proto.Empty{})
	if err != nil {
		logger.Warning(err)
		return
	}
	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		select {
		case <-pc.ctx.Done():
			return
		case <-ctx.Done():
			return
		default:
			resp, err := pdc.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				logger.Error(err)
				time.Sleep(time.Second)
				continue
			}
			wg.Add(1)
			go func(d *proto.Device) {
				defer wg.Done()
				// 设备是否在线
				if !pc.IsOnline(d.Identity) {
					return
				}

				device := DiscoverResponse{
					Identity:     d.Identity,
					Model:        d.Model,
					Manufacturer: d.Manufacturer,
					Name:         fmt.Sprintf("%s_%s", d.Manufacturer, d.Model),
					PluginID:     pc.pluginID,
					AuthRequired: d.AuthRequired,
				}
				select {
				case out <- device:
				default:
				}
			}(resp)

		}
	}
}

// IsOnline 设备是否在线
// 根据缓存判断是否在线，不在线则同步发起请求
func (pc *pluginClient) IsOnline(identity string) bool {
	// 每个设备都开启一个goroutine持续判断是否在线，防止并发请求导致运行多个goroutine
	_, loaded := pc.healthCheckDevice.LoadOrStore(identity, nil)
	if !loaded {
		go pc.healthChecking(identity)
	}

	// 在线则直接返回，不在线则阻塞等待到恢复在线或超时
	if pc.isOnline(identity) {
		return true
	}

	timeout := time.NewTimer(time.Second * 10)
	defer timeout.Stop()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeout.C:
			return false
		case <-ticker.C:
			if err := pc.healthCheck(identity); err != nil {
				logger.Errorf("%s health check err: %s", identity, err.Error())
				return false
			}
			if pc.isOnline(identity) {
				return true
			}
		}
	}
}

func (pc *pluginClient) healthChecking(identity string) {
	healthCheckTicker := time.NewTicker(healthCheckInterval)
	for {
		select {
		case <-pc.ctx.Done():
			logger.Debugf("%s health check done", identity)
			return
		case <-healthCheckTicker.C:
			if err := pc.healthCheck(identity); err != nil {
				logger.Errorf("%s health check err: %s", identity, err.Error())
			}
		}
	}
}

// healthCheck 查看设备的在线状态
func (pc *pluginClient) healthCheck(identity string) error {

	req := proto.HealthCheckReq{Identity: identity}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := pc.protoClient.HealthCheck(ctx, &req)
	if err != nil {
		return err
	}
	if !pc.isOnline(identity) && resp.Online {
		logger.Infof("%s %s back to online", pc.pluginID, identity)
	}
	if resp.Online {
		pc.deviceLastOnlineTime.Store(identity, time.Now())
	} else {
		logger.Debugf("%s %s health check offline", pc.pluginID, identity)
	}
	return nil
}

func (pc *pluginClient) isOnline(identity string) bool {
	if v, ok := pc.deviceLastOnlineTime.Load(identity); ok {
		lastOnlineTime := v.(time.Time)
		return lastOnlineTime.Add(offlineTimeout).After(time.Now())
	}
	return false
}

// GetPluginConfig 获取插件配置
func GetPluginConfig(addr, pluginID string) (config Plugin, err error) {
	url := fmt.Sprintf("http://%s/api/plugin/%s/config.json", addr, pluginID)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(data, &config); err != nil {
		return
	}
	return
}
