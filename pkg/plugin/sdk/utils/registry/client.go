package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

const (
	registerTTL = 10

	etcdURL = "http://0.0.0.0:2379"

	managerTarget = "/sa/plugins"
)

func endpointsTarget(service string) string {
	return fmt.Sprintf("%s/%s", managerTarget, service)
}

// RegisterService 注册服务
func RegisterService(ctx context.Context, service, addr string) {
	logrus.Infoln("register service:", service, addr)
	cli, err := clientv3.NewFromURL(etcdURL)
	if err != nil {
		logrus.Errorf("new etcd client err: %s", err.Error())
		return
	}
	defer cli.Close()
	for {
		if err = register(ctx, cli, service, addr); err != nil {
			logrus.Errorf("register service err: %s", err.Error())
		}
		time.Sleep(time.Second)
	}
}

func register(ctx context.Context, cli *clientv3.Client, service, addr string) (err error) {
	em, err := endpoints.NewManager(cli, managerTarget)
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

	err = em.AddEndpoint(ctx,
		endpointsTarget(service),
		endpoints.Endpoint{Addr: addr},
		clientv3.WithLease(resp.ID),
	)
	if err != nil {
		return
	}
	for {
		if _, ok := <-kl; !ok {
			time.Sleep(time.Second)
			return register(ctx, cli, service, addr)
		}
	}
}

// UnregisterService 取消注册服务
func UnregisterService(ctx context.Context, service string) (err error) {
	logrus.Infoln("unregister service:", service)
	cli, err := clientv3.NewFromURL(etcdURL)
	if err != nil {
		return
	}
	em, err := endpoints.NewManager(cli, managerTarget)
	if err != nil {
		return
	}

	return em.DeleteEndpoint(ctx, endpointsTarget(service))
}
