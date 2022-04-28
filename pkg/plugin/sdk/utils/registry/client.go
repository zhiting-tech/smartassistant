package registry

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"

	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

const (
	registerTTL = 10

	etcdURL = "http://0.0.0.0:2379"

	ManagerTarget = "/sa/plugins"
)

func EndpointsKey(service string) string {
	return fmt.Sprintf("%s/%s", ManagerTarget, service)
}

// RegisterService 注册服务
func RegisterService(ctx context.Context, key string, endpoint endpoints.Endpoint) {
	logger.Info("register service:", key, endpoint.Addr)
	cli, err := clientv3.NewFromURL(etcdURL)
	if err != nil {
		logger.Errorf("new etcd client err: %s", err.Error())
		return
	}
	defer cli.Close()
	for {
		if err = register(ctx, cli, key, endpoint); err != nil {
			logger.Errorf("register service err: %s", err.Error())
		}
		time.Sleep(time.Second)
	}
}

func register(ctx context.Context, cli *clientv3.Client, key string, endpoint endpoints.Endpoint) (err error) {
	em, err := endpoints.NewManager(cli, ManagerTarget)
	if err != nil {
		return
	}

	lease := clientv3.NewLease(cli)
	resp, err := lease.Grant(ctx, registerTTL)
	if err != nil {
		return
	}
	kl, err := lease.KeepAlive(ctx, resp.ID)
	if err != nil {
		return
	}

	err = em.AddEndpoint(ctx, key, endpoint, clientv3.WithLease(resp.ID))
	if err != nil {
		return
	}
	for {
		if _, ok := <-kl; !ok {
			time.Sleep(time.Second)
			return register(ctx, cli, key, endpoint)
		}
	}
}

// UnregisterService 取消注册服务
func UnregisterService(ctx context.Context, key string) (err error) {
	logger.Info("unregister service:", key)
	cli, err := clientv3.NewFromURL(etcdURL)
	if err != nil {
		return
	}
	em, err := endpoints.NewManager(cli, ManagerTarget)
	if err != nil {
		return
	}

	return em.DeleteEndpoint(ctx, key)
}
