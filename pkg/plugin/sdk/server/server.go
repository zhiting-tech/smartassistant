package server

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/archive"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto"
	"math/rand"
	"os"
	"time"
)

type Server struct {
	Manager      *Manager
	Domain       string
	Router       *gin.Engine
	ApiRouter    *gin.RouterGroup
	pluginRouter *gin.RouterGroup
	configFile   string
	staticDir    string
}

func (p Server) OTA(req *proto.OTAReq, server proto.Plugin_OTAServer) error {
	logrus.Debugf("%s OTA with firmware url %s", req.Identity, req.FirmwareUrl)
	ch, err := p.Manager.OTA(req.Identity, req.FirmwareUrl)
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
				Identity: req.Identity,
				Step:     int32(v.Step),
			}
			if err = server.Send(&resp); err != nil {
				logrus.Errorf("send ota response error: %s", err.Error())
			}
			if v.Step == OTAFinish {
				return nil
			}
		}
	}
}

func (p Server) HealthCheck(context context.Context, req *proto.HealthCheckReq) (resp *proto.HealthCheckResp, err error) {

	online := p.Manager.HealthCheck(req.Identity)
	logrus.Debugf("%s HealthCheck,online: %v", req.Identity, online)
	resp = &proto.HealthCheckResp{
		Identity: req.Identity,
		Online:   online,
	}
	return
}

func (p Server) Discover(request *proto.Empty, server proto.Plugin_DiscoverServer) error {
	devices, _ := p.Manager.Devices()
	for _, device := range devices {
		d := proto.Device{
			Identity:     device.Identity(),
			Model:        device.Info().Model,
			Manufacturer: device.Info().Manufacturer,
		}
		_, d.AuthRequired = device.(AuthDevice)
		server.Send(&d)
	}
	return nil
}

func (p Server) Connect(ctx context.Context, req *proto.AuthReq) (resp *proto.GetAttributesResp, err error) {
	logrus.Debugf("%s connect with auth params %v", req.Identity, req.Params)
	if err = p.Manager.Auth(req.Identity, req.Params); err != nil {
		return
	}

	// getAttrsReq := proto.GetAttributesReq{Identity: req.Identity}
	// return p.GetAttributes(ctx, &getAttrsReq)

	return &proto.GetAttributesResp{}, nil
}

func (p Server) Disconnect(ctx context.Context, req *proto.AuthReq) (resp *proto.Empty, err error) {
	logrus.Debugf("%s disconnect with params %v", req.Identity, req.Params)
	if err = p.Manager.Disconnect(req.Identity, req.Params); err != nil {
		return
	}
	return
}

func (p Server) GetAttributes(context context.Context, request *proto.GetAttributesReq) (resp *proto.GetAttributesResp, err error) {
	logrus.Debugf("%s GetAttributes", request.Identity)

	instances, err := p.Manager.GetAttributes(request.Identity)
	if err != nil {
		return
	}

	resp = new(proto.GetAttributesResp)
	resp.Success = true
	for _, instance := range instances {

		data, _ := json.Marshal(instance.Attributes)
		ins := proto.Instance{
			Type:       instance.Type,
			Identity:   request.Identity,
			InstanceId: int32(instance.InstanceId),
			Attributes: data,
		}
		resp.Instances = append(resp.Instances, &ins)
	}
	resp.OtaSupport, err = p.Manager.IsOTASupport(request.Identity)
	if err != nil {
		return
	}
	logrus.Println("instances resp:", resp)
	return
}

type SetAttribute struct {
	InstanceID int         `json:"instance_id"`
	Attribute  string      `json:"attribute"`
	Val        interface{} `json:"val"`
}

type SetRequest struct {
	Attributes []SetAttribute `json:"attributes"`
}

func (p Server) SetAttributes(context context.Context, request *proto.SetAttributesReq) (resp *proto.SetAttributesResp, err error) {
	logrus.Debugf("%s SetAttribute", request.Identity)

	var req SetRequest
	err = json.Unmarshal(request.Data, &req)
	if err != nil {
		return
	}
	for _, attr := range req.Attributes {
		logrus.Debugf("set %s %d attr %s %v:\n", request.Identity, attr.InstanceID, attr.Attribute, attr.Val)
		err = p.Manager.SetAttribute(request.Identity, attr.InstanceID, attr.Attribute, attr.Val)
		if err != nil {
			return
		}
	}
	resp = new(proto.SetAttributesResp)
	resp.Success = true
	return
}
func (p Server) StateChange(request *proto.Empty, server proto.Plugin_StateChangeServer) error {
	logrus.Println("stateChange requesting...")

	nc := make(chan Notify, 20)

	p.Manager.Subscribe(nc)
	defer p.Manager.Unsubscribe(nc)

	for {
		select {
		case <-server.Context().Done():
			return nil
		case n := <-nc:
			var s proto.State
			s.Identity = n.Identity
			s.InstanceId = int32(n.InstanceID)
			s.Attributes, _ = json.Marshal(n.Attribute)
			logrus.Debugf("notification:%s\n", s.Attributes)
			server.Send(&s)
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

func NewPluginServer(opts ...OptionFunc) *Server {
	m := NewManager()
	m.Init()

	domain := os.Getenv("PLUGIN_DOMAIN")
	if domain == "" {
		bytes := make([]byte, 4)
		rand.Seed(time.Now().UnixNano())
		rand.Read(bytes)
		domain = hex.EncodeToString(bytes)
	}

	route := gin.Default()
	path := fmt.Sprintf("api/plugin/%s", domain)
	pluginGroup := route.Group(path)

	s := Server{
		Manager:      m,
		Domain:       domain,
		Router:       route,
		pluginRouter: pluginGroup,
		ApiRouter:    pluginGroup.Group("api"),
		staticDir:    "./html",
		configFile:   "./config.json",
	}
	for _, opt := range opts {
		opt(&s)
	}
	s.Init()
	return &s
}
