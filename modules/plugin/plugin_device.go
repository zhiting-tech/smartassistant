package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	event2 "github.com/zhiting-tech/smartassistant/pkg/event"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto/v2"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

type device struct {
	iid         string
	areaID      uint64
	pluginID    string
	protoClient proto.PluginClient // 请求插件服务的grpc客户端

	once           sync.Once
	lastOnlineTime time.Time

	closed chan struct{}
}

// WaitOnline 等待直到设备在线
func (d *device) WaitOnline(ctx context.Context) error {

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second*30)
		defer cancel()
	}
	ticker := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case <-ticker.C:
			if d.IsOnline() {
				return nil
			}
		case <-ctx.Done():
			return errors.New("wait online timeout")
		}
	}
}

func (d *device) IsOnline() bool {
	return d.lastOnlineTime.Add(offlineTimeout).After(time.Now())
}

func (d *device) HealthCheck() {
	d.once.Do(
		func() {
			go d.healthChecking()
		})
}

func (d *device) healthChecking() {
	logger.Debugf("%s start health check", d.iid)

	// 马上发起一次请求
	if err := d.healthCheck(); err != nil {
		logger.Errorf("%s health check err: %s", d.iid, err.Error())
	}
	healthCheckTicker := time.NewTicker(healthCheckInterval)
	for {
		select {
		case <-d.closed:
			logger.Debugf("%s health check done", d.iid)
			return
		case <-healthCheckTicker.C:
			if err := d.healthCheck(); err != nil {
				logger.Errorf("%s health check err: %s", d.iid, err.Error())
			}
		}
	}

}

// healthCheck 查看设备的在线状态
func (d *device) healthCheck() error {

	req := proto.HealthCheckReq{Iid: d.iid}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, err := d.protoClient.HealthCheck(ctx, &req)
	if err != nil {
		return err
	}

	em := event2.NewEventMessage(event2.OnlineStatus, d.areaID)
	em.Param = map[string]interface{}{
		"plugin_id": d.pluginID,
		"iid":       d.iid,
		"online":    resp.Online,
	}
	if !d.IsOnline() && resp.Online { // online
		logger.Infof("%s %s back to online", d.pluginID, d.iid)
		event2.Notify(em)
	}

	if d.IsOnline() && !resp.Online { // offline
		logger.Infof("%s %s offline", d.pluginID, d.iid)
		event2.Notify(em)
	}

	if resp.Online {
		d.lastOnlineTime = time.Now()
	}

	return nil
}

func (d *device) Connect(ctx context.Context, authParams map[string]interface{}) (das thingmodel.ThingModel, err error) {
	params, _ := json.Marshal(authParams)
	req := proto.AuthReq{
		Iid:    d.iid,
		Params: params,
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5) // 配对认证过程较长可能长达数分钟
	defer cancel()

	resp, err := d.protoClient.Connect(ctx, &req)
	if err != nil {
		return
	}
	das = ParseAttrsResp(resp)
	logger.Debugf("connect resp: %#v\n", das)
	return
}

func (d *device) Disconnect(ctx context.Context, authParams map[string]interface{}) (err error) {
	params, _ := json.Marshal(authParams)
	req := proto.AuthReq{
		Iid:    d.iid,
		Params: params,
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	_, err = d.protoClient.Disconnect(ctx, &req)
	if err != nil {
		return
	}
	return d.Close()
}

func (d *device) GetAttributes(ctx context.Context) (das thingmodel.ThingModel, err error) {
	req := proto.GetInstancesReq{Iid: d.iid}
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	resp, err := d.protoClient.GetInstances(ctx, &req)
	if err != nil {
		return
	}
	// logger.Debugf("GetInstances resp: %#v\n", resp)
	das = ParseAttrsResp(resp)
	return
}

func (d *device) OTA(ctx context.Context, firmwareURL string) (err error) {
	req := proto.OTAReq{
		Iid:         d.iid,
		FirmwareUrl: firmwareURL,
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()
	otaCli, err := d.protoClient.OTA(ctx, &req)
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
		logger.Println("ota response:", resp.Iid, resp.Step)
		if resp.Step >= 100 {
			return nil
		}
		if resp.Step < 0 {
			return fmt.Errorf("ota err, step: %d", resp.Step)
		}
	}
}

func (d *device) Close() (err error) {
	d.closed <- struct{}{}
	close(d.closed)
	return nil
}
