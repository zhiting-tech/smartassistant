package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	event2 "github.com/zhiting-tech/smartassistant/pkg/event"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto/v2"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"

	"google.golang.org/protobuf/types/known/emptypb"
)

var NotExistErr = errors.New("plugin not exist")

func NewClient() *client {
	return &client{
		clients: make(map[string]*pluginClient),
	}
}

type client struct {
	mu      sync.Mutex // clients 锁
	clients map[string]*pluginClient

	devicesCancel sync.Map
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

func (c *client) DeviceConfigs() (configs []DeviceConfig) {

	for _, cli := range c.clients {
		for _, d := range cli.PluginConf.SupportDevices {
			d.PluginID = cli.pluginID
			configs = append(configs, d)
		}
	}
	return
}

func (c *client) DeviceConfig(pluginID, model string) (config DeviceConfig) {
	cli, err := c.get(pluginID)
	if err != nil {
		return
	}

	for _, sd := range cli.PluginConf.SupportDevices {
		if model != sd.Model {
			continue
		}
		sd.PluginID = cli.pluginID
		return sd
	}
	return
}

// Connect 连接设备
func (c *client) Connect(ctx context.Context, identify Identify, authParams map[string]interface{}) (das thingmodel.ThingModel, err error) {

	pc, err := c.get(identify.PluginID)
	if err != nil {
		return
	}
	return pc.Connect(ctx, identify.IID, authParams)
}

func (c *client) Disconnect(ctx context.Context, identify Identify, authParams map[string]interface{}) (err error) {
	cli, err := c.get(identify.PluginID)
	if err != nil {
		return
	}
	err = cli.Disconnect(ctx, identify.IID, authParams)
	if err != nil {
		return
	}

	v, loaded := c.devicesCancel.LoadAndDelete(identify.IID)
	if loaded {
		if cancel, ok := v.(context.CancelFunc); ok {
			cancel()
		}
	}
	return nil
}

func (c *client) GetAttributes(ctx context.Context, identify Identify) (das thingmodel.ThingModel, err error) {
	req := proto.GetInstancesReq{Iid: identify.IID}
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	cli, err := c.get(identify.PluginID)
	if err != nil {
		return
	}
	resp, err := cli.protoClient.GetInstances(ctx, &req)
	if err != nil {
		return
	}
	// logger.Debugf("GetInstances resp: %#v\n", resp)
	das = ParseAttrsResp(resp)
	return
}

func (c *client) SetAttributes(ctx context.Context, pluginID string, areaID uint64, setReq sdk.SetRequest) (result []byte, err error) {
	data, _ := json.Marshal(setReq)
	req := proto.SetAttributesReq{
		Data: data,
	}
	logger.Debug("set attributes: ", string(data))
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	cli, err := c.get(pluginID)
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

func (c *client) IsOnline(identify Identify) bool {
	cli, err := c.get(identify.PluginID)
	if err != nil {
		logger.Warningf("plugin %s not found", identify.PluginID)
		return false
	}
	return cli.IsOnline(identify.IID)
}

func (c *client) OTA(ctx context.Context, identify Identify, firmwareURL string) (err error) {
	req := proto.OTAReq{
		Iid:         identify.IID,
		FirmwareUrl: firmwareURL,
	}
	logger.Debugf("ota: %s, firmware url: %s", identify.IID, firmwareURL)
	ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()
	cli, err := c.get(identify.PluginID)
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
		logger.Println("ota response:", resp.Iid, resp.Step)
		if resp.Step >= 100 {
			return nil
		}
		if resp.Step < 0 {
			return fmt.Errorf("ota err, step: %d", resp.Step)
		}
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
	go cli.InitDevices()
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

func (c *client) ListenStateChange(pluginID string) {
	cli, err := c.get(pluginID)
	if err != nil {
		return
	}
	pdc, err := cli.protoClient.Subscribe(cli.ctx, &emptypb.Empty{})
	if err != nil {
		logger.Error("state onDeviceStateChange error:", err)
		return
	}
	logger.Println("StateChange recv...")
	for {
		var resp *proto.Event
		resp, err = pdc.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Println(err)
			// TODO retry
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

			if err = HandleEvent(cli, ev); err != nil {
				logger.Errorf("handle event err:%s", err)
			}
		}()
	}
	logger.Println("StateChangeFromPlugin exit")
}

func HandleEvent(cli *pluginClient, ev sdk.Event) (err error) {
	switch ev.Type {
	case "attr_change":

		var d entity.Device
		var attrEvent definer.AttributeEvent
		_ = json.Unmarshal(ev.Data, &attrEvent)

		d, err = entity.GetPluginDevice(cli.areaID, cli.pluginID, attrEvent.IID)
		if err != nil {
			logger.Errorf("GetPluginDevice error:%s", err.Error())
			return
		}

		em := event2.NewEventMessage(event2.AttributeChange, cli.areaID)
		em.SetDeviceID(d.ID)
		em.SetAttr(attrEvent)
		event2.Notify(em)
	case "thing_model_change":
		var tme definer.ThingModelEvent
		_ = json.Unmarshal(ev.Data, &tme)
		em := event2.NewEventMessage(event2.ThingModelChange, cli.areaID)
		em.Param = map[string]interface{}{
			"thing_model": tme.ThingModel,
			"iid":         tme.IID,
			"area_id":     cli.areaID,
			"plugin_id":   cli.pluginID,
		}
		event2.Notify(em)
	}
	return
}

func ParseAttrsResp(resp *proto.GetInstancesResp) thingmodel.ThingModel {

	var instances []thingmodel.Instance
	_ = json.Unmarshal(resp.Instances, &instances)
	return thingmodel.ThingModel{
		Instances:  instances,
		OTASupport: resp.OtaSupport,
	}
}
