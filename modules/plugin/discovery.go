package plugin

import (
	"context"
	"strings"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/reverseproxy"

	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

const (
	etcdURL = "http://etcd:2379"

	managerTarget = "/sa/plugins"
)

func EndPointsManager() (manager endpoints.Manager, err error) {

	cli, err := clientv3.NewFromURL(etcdURL)
	if err != nil {
		return
	}
	em, err := endpoints.NewManager(cli, managerTarget)
	if err != nil {
		return
	}

	return em, nil
}

type discovery struct {
	client *client
}

func NewDiscovery(cli *client) *discovery {
	return &discovery{
		client: cli,
	}
}

// Listen 监听etcd发现服务
func (m *discovery) Listen(ctx context.Context) (err error) {
	logger.Println("start discovering service")
	em, err := EndPointsManager()
	if err != nil {
		logger.Error("get endpoint manager err:", err.Error())
		return
	}

	// watch etcd service onDeviceStateChange
	w, err := em.NewWatchChannel(ctx)
	if err != nil {
		logger.Error("new watch channel err:", err.Error())
		return
	}

	for updates := range w {
		if err = m.handleUpdates(updates); err != nil {
			logger.Error("handle update err:", err.Error())
		}
	}
	return
}
func (m *discovery) handleUpdates(updates []*endpoints.Update) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("handleUpdates panic: %v", r)
		}
	}()

	for _, update := range updates {
		switch update.Op {
		case endpoints.Delete:
			if err = m.unregisterService(update.Key); err != nil {
				logger.Error("unregister service err:", err.Error())
			}
		case endpoints.Add:
			if err = m.registerService(update.Key, update.Endpoint); err != nil {
				logger.Error("register service err:", err.Error())
			}
		}
	}
	return
}

// registerService 注册插件服务(grpc和http)
func (m *discovery) registerService(key string, endpoint endpoints.Endpoint) error {

	service := strings.TrimPrefix(key, managerTarget+"/")
	logger.Debugf("register service %s:%s from etcd", service, endpoint.Addr)

	// FIXME 插件暂时使用host模式，直接访问宿主机地址
	// endpoint.Addr = config.GetConf().SmartAssistant.HostIP
	if err := reverseproxy.RegisterUpstream(service, endpoint.Addr); err != nil {
		return err
	}

	plgConf, err := GetPluginConfig(endpoint.Addr, service)
	if err != nil {
		return err
	}
	// FIXME 仅支持单个家庭
	area, err := getCurrentArea()
	if err != nil {
		logger.Errorf("getCurrentArea err: %s", err.Error())
		return err
	}
	cli, err := newClient(area.ID, service, key, plgConf)
	if err != nil {
		logger.Errorf("new client err: %s", err.Error())
		return err
	}
	m.client.Add(cli)
	return nil
}

// unregisterService 取消插件注册服务
func (m *discovery) unregisterService(key string) error {

	service := strings.TrimPrefix(key, managerTarget+"/")
	logger.Debugf("unregister service %s from etcd", service)
	if err := reverseproxy.UnregisterUpstream(service); err != nil {
		return err
	}
	if err := m.client.Remove(service); err != nil {
		return err
	}
	return nil
}

// getCurrentArea 获取当前家庭
func getCurrentArea() (area entity.Area, err error) {
	if err = entity.GetDB().First(&area).Error; err != nil {
		return
	}
	return
}
