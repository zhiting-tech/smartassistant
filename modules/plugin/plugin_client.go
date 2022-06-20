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
	plgInfo, err := entity.GetPlugin(plgID, areaID)
	if err != nil {
		return nil, err
	}
	return &pluginClient{
		areaID:      areaID,
		pluginID:    plgID,
		PluginInfo:  plgInfo,
		protoClient: proto.NewPluginClient(conn),
		PluginConf:  plgConf,
	}, nil
}

type pluginClient struct {
	areaID        uint64
	pluginID      string
	PluginInfo    entity.PluginInfo
	protoClient   proto.PluginClient // 请求插件服务的grpc客户端
	PluginConf    Plugin
	endpointsMeta sdk.MetaData

	devices sync.Map
}

func (pc *pluginClient) Device(iid string) *device {

	newDevice := &device{
		iid:         iid,
		areaID:      pc.areaID,
		pluginID:    pc.pluginID,
		protoClient: pc.protoClient,
		closed:      make(chan struct{}),
	}
	if v, loaded := pc.devices.LoadOrStore(iid, newDevice); loaded {
		if d, ok := v.(*device); ok {
			return d
		}
	}
	newDevice.HealthCheck()
	return newDevice
}

func (pc *pluginClient) RemoveDevice(ctx context.Context, iid string, authParams map[string]interface{}) error {

	if v, loaded := pc.devices.LoadAndDelete(iid); loaded {
		if d, ok := v.(*device); ok {
			return d.Disconnect(ctx, authParams)
		}
	}
	return nil
}

func (pc *pluginClient) Stop() {
	pc.devices.Range(func(key, value interface{}) bool {
		d, ok := value.(*device)
		if ok {
			d.Close()
		}
		pc.devices.Delete(key)
		return true
	})
}

func (pc *pluginClient) DeviceDiscover(ctx context.Context, out chan<- DiscoverResponse) {

	pdc, err := pc.protoClient.Discover(ctx, &emptypb.Empty{})
	if err != nil {
		logger.Warning(err)
		return
	}
	for {
		select {
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
				Type:         resp.Type,
				Name:         pc.GetDeviceName(resp.Model),
				PluginID:     pc.pluginID,
				PluginName:   pc.PluginInfo.Name,
				AuthRequired: resp.AuthRequired,
			}

			if device.Name == "" {
				logger.Debugf("device %s name is empty", device.IID)
				continue
			}

			if resp.AuthRequired {
				var authParams []thingmodel.AuthParam
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

func (pc *pluginClient) InitDevices() {
	devices, err := entity.GetDevicesByPluginID(pc.pluginID)
	if err != nil {
		logger.Errorf("get devices err when init devices: %s", err.Error())
		return
	}

	for _, d := range devices {
		go pc.Device(d.IID) // 通过触发healthCheck来建立连接
	}
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

func (pc *pluginClient) ListenChange() error {
	pdc, err := pc.protoClient.Subscribe(context.Background(), &emptypb.Empty{})
	if err != nil {
		return err
	}
	logger.Println("StateChange recv...")
	for {
		var resp *proto.Event
		resp, err = pdc.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error(r)
				}
			}()
			var ev sdk.Event
			_ = json.Unmarshal(resp.Data, &ev)

			if err = HandleEvent(pc, ev); err != nil {
				logger.Errorf("handle event err:%s", err)
			}
		}()
	}
	logger.Println("StateChangeFromPlugin exit")
	return err
}
