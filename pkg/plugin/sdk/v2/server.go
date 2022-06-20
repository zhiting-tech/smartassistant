package sdk

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/zhiting-tech/smartassistant/pkg/archive"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto/v2"
	"github.com/zhiting-tech/smartassistant/pkg/trace"
)

type Server struct {
	Manager      *Manager
	Domain       string
	Router       *gin.Engine
	ApiRouter    *gin.RouterGroup
	pluginRouter *gin.RouterGroup
	configFile   string
	staticDir    string
	discoverFunc DiscoverFunc
}

func (p Server) OTA(req *proto.OTAReq, server proto.Plugin_OTAServer) error {
	logrus.Debugf("%s OTA with firmware url %s", req.Iid, req.FirmwareUrl)
	ch, err := p.Manager.OTA(req.Iid, req.FirmwareUrl)
	if err != nil {
		return err
	}

	timeout := time.NewTimer(time.Minute * 10)

	for {
		select {
		case <-timeout.C:
			return errors.New("OTA timeout")
		case v, ok := <-ch:
			if !ok {
				return nil
			}
			resp := proto.OTAResp{
				Iid:  req.Iid,
				Step: int32(v.Step),
			}
			if err = server.Send(&resp); err != nil {
				logrus.Errorf("send ota response error: %s", err.Error())
			}
		}
	}
}

func (p Server) HealthCheck(context context.Context, req *proto.HealthCheckReq) (resp *proto.HealthCheckResp, err error) {

	online := p.Manager.HealthCheck(req.Iid)
	resp = &proto.HealthCheckResp{
		Iid:    req.Iid,
		Online: online,
	}
	return
}

func (p Server) Discover(request *emptypb.Empty, server proto.Plugin_DiscoverServer) error {

	devices, err := p.Manager.Devices()
	if err != nil {
		return err
	}
	for _, d := range devices {
		ad, authRequired := d.(AuthDevice)
		pd := proto.Device{
			Iid:          d.Info().IID,
			Model:        d.Info().Model,
			Manufacturer: d.Info().Manufacturer,
			Type:         d.Info().Type,
			AuthRequired: authRequired,
		}
		if authRequired && len(ad.AuthParams()) != 0 {
			pd.AuthParams, _ = json.Marshal(ad.AuthParams())
		}
		server.Send(&pd)
	}
	return nil
}

func (p Server) Connect(ctx context.Context, req *proto.AuthReq) (resp *proto.GetInstancesResp, err error) {
	logrus.Debugf("%s connect with auth params %v", req.Iid, req.Params)

	var params map[string]interface{}
	json.Unmarshal(req.Params, &params)

	d, err := p.Manager.GetDevice(req.Iid)
	if err != nil {
		if !errors2.Is(err, DeviceNotExist) {
			return
		}
		// 找不到设备时，阻塞等待发现设备并初始化
		if err = p.discoverDevice(ctx, req.Iid); err != nil {
			return
		}
	}
	if err = p.Manager.Connect(d, params); err != nil {
		return
	}

	getAttrsReq := proto.GetInstancesReq{Iid: req.Iid}
	return p.GetInstances(ctx, &getAttrsReq)
}

func (p Server) Disconnect(ctx context.Context, req *proto.AuthReq) (resp *emptypb.Empty, err error) {
	logrus.Debugf("%s disconnect with params %v", req.Iid, req.Params)
	resp = new(emptypb.Empty)
	var params map[string]interface{}
	json.Unmarshal(req.Params, &params)
	if err = p.Manager.Disconnect(req.Iid, params); err != nil {
		return
	}
	return
}

func (p Server) GetInstances(context context.Context, request *proto.GetInstancesReq) (resp *proto.GetInstancesResp, err error) {
	logrus.Debugf("%s GetAttribute", request.Iid)

	tm, err := p.Manager.GetThingModel(request.Iid)
	if err != nil {
		return
	}

	resp = new(proto.GetInstancesResp)
	resp.Success = true

	logrus.Println(tm.Instances)

	resp.Instances, err = json.Marshal(tm.Instances)
	if err != nil {
		logrus.Errorf("newlisht err: %s", err.Error())
		return
	}
	resp.OtaSupport, err = p.Manager.IsOTASupport(request.Iid)
	if err != nil {
		return
	}
	d, err := p.Manager.getDevice(request.Iid)
	if err != nil {
		return
	}
	var ad AuthDevice
	ad, resp.AuthRequired = d.(AuthDevice)
	if resp.AuthRequired {
		resp.IsAuth = ad.IsAuth()
		resp.AuthParams, _ = json.Marshal(ad.AuthParams())
	}
	logrus.Println("instances resp:", resp)
	return
}

type SetAttribute struct {
	IID string      `json:"iid"`
	AID int         `json:"aid"`
	Val interface{} `json:"val"`
}

type SetRequest struct {
	Attributes []SetAttribute `json:"attributes"`
}

func (p Server) SetAttributes(context context.Context, request *proto.SetAttributesReq) (resp *proto.SetAttributesResp, err error) {
	logrus.Debugf("%v SetAttribute", request)

	var req SetRequest
	err = json.Unmarshal(request.Data, &req)
	if err != nil {
		return
	}
	err = p.Manager.SetAttributes(req.Attributes)
	if err != nil {
		return
	}
	resp = new(proto.SetAttributesResp)
	resp.Success = true
	return
}

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type EventChan chan Event

func (p Server) Subscribe(request *emptypb.Empty, server proto.Plugin_SubscribeServer) error {
	logrus.Println("stateChange requesting...")

	nc := make(EventChan, 20)

	p.Manager.Subscribe(nc)
	defer p.Manager.Unsubscribe(nc)
	for {
		select {
		case <-server.Context().Done():
			return nil
		case n := <-nc:
			var s proto.Event
			s.Data, _ = json.Marshal(n)
			logrus.Printf("notification: %s\n", s.Data)
			server.Send(&s)
		}
	}
}

func (p *Server) discovering() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	go p.discover(ctx)
	ticker := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-ticker.C:
			cancel() // 停止发现
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*20)
			go p.discover(ctx)
		}
	}
}

func (p *Server) Init() {
	p.pluginRouter.Group("html").Static("", p.staticDir)
	p.pluginRouter.StaticFile("config.json", p.configFile)

	// 压缩静态文件，返回压缩包
	fileName := fmt.Sprintf("%s.zip", p.Domain)

	if !Exist(fileName) {
		if err := archive.Zip(fileName, p.staticDir, p.configFile); err != nil {
			logrus.Errorf("archive file %s err: %s", p.staticDir, err.Error())
			return
		}
	}
	archiveAPI := fmt.Sprintf("resources/archive/%s", fileName)
	p.pluginRouter.StaticFile(archiveAPI, fileName)

}

func Exist(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	// if errors.Is(err, os.ErrNotExist) {
	//	return false, nil
	// }
	return false
}

type OptionFunc func(s *Server)

func WithStatic(staticDir string) OptionFunc {
	return func(s *Server) {
		s.staticDir = staticDir
	}
}
func WithConfigFile(configFile string) OptionFunc {
	return func(s *Server) {
		s.configFile = configFile
	}
}
func WithDomain(domain string) OptionFunc {
	return func(s *Server) {
		s.Domain = domain
	}
}

func NewPluginServer(discoverFunc DiscoverFunc, opts ...OptionFunc) *Server {
	m := NewManager()
	m.Init()

	domain := os.Getenv("PLUGIN_DOMAIN")
	if domain == "" {
		bytes := make([]byte, 4)
		rand.Seed(time.Now().UnixNano())
		rand.Read(bytes)
		domain = hex.EncodeToString(bytes)
	}
	trace.Init(domain)
	route := gin.New()
	route.Use(gin.Recovery())
	path := fmt.Sprintf("api/plugin/%s", domain)
	pluginGroup := route.Group(path)
	apiGroup := pluginGroup.Group("api")
	apiGroup.Use(gin.Logger())

	s := Server{
		discoverFunc: discoverFunc,
		Manager:      m,
		Domain:       domain,
		Router:       route,
		pluginRouter: pluginGroup,
		ApiRouter:    apiGroup,
		staticDir:    "./html",
		configFile:   "./config.json",
	}
	for _, opt := range opts {
		opt(&s)
	}
	s.Init()
	return &s
}

// discover 发现设备并刷新发现设备列表
func (p *Server) discover(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("discovering panic: %v", r)
		}
	}()

	devices := make(chan Device, 10)
	go func() {
		logrus.Debug("discovering...")
		p.discoverFunc(ctx, devices)
		logrus.Debug("discovering done")
		close(devices)
	}()

	var dd []Device
	for d := range devices {
		if err := p.Manager.InitOrUpdateDevice(d); err != nil {
			logrus.Errorf("init or update device err %s", err.Error())
			return
		}
		dd = append(dd, d)
	}
	p.Manager.discoveredDevices = make([]Device, len(dd))
	copy(p.Manager.discoveredDevices, dd)

	logrus.Debugf("discover %d devices", len(p.Manager.discoveredDevices))
}

// discoverDevice 发现设备
func (p *Server) discoverDevice(ctx context.Context, iid string) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("discovering device panic: %v", r)
		}
	}()

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second*10)
		defer cancel()
	}

	devices := make(chan Device, 10)
	go func() {
		defer close(devices)
		logrus.Debugf("discovering %s...", iid)
		p.discoverFunc(ctx, devices)
		logrus.Debugf("discover %s done", iid)
	}()

	for {
		select {
		case <-time.NewTimer(time.Second * 10).C:
			return DeviceNotExist
		case d, ok := <-devices:
			if !ok {
				logrus.Debugf("discover %s done, not found", iid)
				return DeviceNotExist
			}
			if d.Info().IID == iid {
				if err := p.Manager.InitOrUpdateDevice(d); err != nil {
					logrus.Errorf("init or update device err %s", err.Error())
					return err
				}
				logrus.Debugf("device %s found and init", iid)
				return nil
			}
		}
	}

}
