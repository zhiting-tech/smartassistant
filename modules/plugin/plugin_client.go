package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/version"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto/v2"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

const (
	healthCheckInterval = time.Second * 10 // 存活检查间隔
	offlineTimeout      = time.Second * 15 // 超过这个时间没有响应在线判断为离线
)

func newClient(areaID uint64, plgID string, endpoint endpoints.Endpoint) (*pluginClient, error) {

	meta, ok := endpoint.Metadata.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid metadata of plugin %s, metadata: %+v", plgID, endpoint.Metadata)
	}
	sdkVersion, _ := meta["sdk_version"].(string)
	if !version.IsValid(sdkVersion) {
		return nil, fmt.Errorf("plugin %s's sdk version %s is invalid", plgID, sdkVersion)
	}
	logger.Debugf("plugin: %s, sdk version: %s", plgID, sdkVersion)
	// TODO 根据 meta.SDKVersion 兼容不同版本的插件sdk

	plgConf, err := GetPluginConfig(endpoint.Addr, plgID)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(
		endpoint.Addr,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
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
	areaID        uint64
	pluginID      string
	protoClient   proto.PluginClient // 请求插件服务的grpc客户端
	cancel        context.CancelFunc
	ctx           context.Context
	PluginConf    Plugin
	endpointsMeta sdk.MetaData

	healthCheckDevice    sync.Map // 记录有存活检查的设备，防止重复发起存活检查
	deviceLastOnlineTime sync.Map // 记录每个设备上次在线时间
}

func (pc *pluginClient) Stop() {
	if pc.cancel != nil {
		pc.cancel()
	}
}

func (pc *pluginClient) DeviceDiscover(ctx context.Context, out chan<- DiscoverResponse) {

	pdc, err := pc.protoClient.Discover(ctx, &emptypb.Empty{})
	if err != nil {
		logger.Warning(err)
		return
	}
	for {
		select {
		case <-pc.ctx.Done():
			return
		case <-ctx.Done():
			return
		default:
			var resp *proto.Device
			resp, err = pdc.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				logger.Error(err)
				time.Sleep(time.Second)
				continue
			}
			logrus.Warningf("discover: (%s/%s/%s)", resp.Manufacturer, resp.Model, resp.Iid)

			device := DiscoverResponse{
				IID:          resp.Iid,
				Model:        resp.Model,
				Manufacturer: resp.Manufacturer,
				Name:         pc.GetDeviceName(resp.Model),
				PluginID:     pc.pluginID,
				AuthRequired: resp.AuthRequired,
			}

			if device.Name == "" {
				logger.Debugf("device %s name is empty", device.IID)
				continue
			}

			if resp.AuthRequired {
				var authParams []sdk.AuthParam
				if err = json.Unmarshal(resp.AuthParams, &authParams); err != nil {
					logrus.Errorf("unmarshal authParams err: %s", err)
					continue
				}
				device.AuthParams = authParams
			}
			out <- device
			logrus.Println("discover sent")
		}
	}
}

// IsOnline 设备是否在线
// 根据缓存判断是否在线，不在线则同步发起请求
func (pc *pluginClient) IsOnline(iid string) bool {

	// 在线则直接返回，不在线则阻塞等待到恢复在线或超时
	if pc.isOnline(iid) {
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
			if err := pc.healthCheck(iid); err != nil {
				logger.Errorf("%s health check err: %s", iid, err.Error())
				return false
			}
			if pc.isOnline(iid) {
				return true
			}
		}
	}

}
func (pc *pluginClient) InitDevices() {
	devices, err := entity.GetDevicesByPluginID(pc.pluginID)
	if err != nil {
		logger.Errorf("get devices err when init devices: %s", err.Error())
		return
	}

	for _, d := range devices {
		go pc.HealthCheck(d.IID)
	}
}

func (pc *pluginClient) HealthCheck(iid string) {

	// 防止并发请求导致运行多个goroutine
	_, loaded := pc.healthCheckDevice.LoadOrStore(iid, nil)
	if loaded {
		return
	}

	logger.Debugf("%s start health check", iid)
	pc.healthChecking(iid)
}

func (pc *pluginClient) healthChecking(iid string) {
	healthCheckTicker := time.NewTicker(healthCheckInterval)
	for {
		select {
		case <-pc.ctx.Done():
			logger.Debugf("%s health check done", iid)
			return
		case <-healthCheckTicker.C:
			if err := pc.healthCheck(iid); err != nil {
				logger.Errorf("%s health check err: %s", iid, err.Error())
			}
		}
	}

}

// healthCheck 查看设备的在线状态
func (pc *pluginClient) healthCheck(iid string) error {

	req := proto.HealthCheckReq{Iid: iid}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, err := pc.protoClient.HealthCheck(ctx, &req)
	if err != nil {
		return err
	}
	if !pc.isOnline(iid) && resp.Online {
		logger.Infof("%s %s back to online", pc.pluginID, iid)
	}

	if resp.Online {
		pc.deviceLastOnlineTime.Store(iid, time.Now())
	} else {
		logger.Debugf("%s %s health check offline", pc.pluginID, iid)
	}

	return nil
}

func (pc *pluginClient) isOnline(iid string) bool {
	if v, ok := pc.deviceLastOnlineTime.Load(iid); ok {
		lastOnlineTime := v.(time.Time)
		return lastOnlineTime.Add(offlineTimeout).After(time.Now())
	}
	return false
}

func (pc *pluginClient) Connect(ctx context.Context, iid string, authParams map[string]interface{}) (das thingmodel.ThingModel, err error) {
	params, _ := json.Marshal(authParams)
	req := proto.AuthReq{
		Iid:    iid,
		Params: params,
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5) // 配对认证过程较长可能长达数分钟
	defer cancel()

	resp, err := pc.protoClient.Connect(ctx, &req)
	if err != nil {
		return
	}
	das = ParseAttrsResp(resp)
	logger.Debugf("connect resp: %#v\n", das)
	return
}

func (pc *pluginClient) Disconnect(ctx context.Context, iid string, authParams map[string]interface{}) (err error) {
	params, _ := json.Marshal(authParams)
	req := proto.AuthReq{
		Iid:    iid,
		Params: params,
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	_, err = pc.protoClient.Disconnect(ctx, &req)
	if err != nil {
		return
	}
	return
}

func (pc *pluginClient) GetDeviceName(model string) string {
	for _, d := range pc.PluginConf.SupportDevices {
		if d.Model == model {
			if d.Name != "" {
				return d.Name
			}
		}
	}
	return model
}
