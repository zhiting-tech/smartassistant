package supervisor

import (
	"context"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/supervisor/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	DefaultSocketAddr = "unix:///mnt/data/zt-smartassistant/run/supervisor/supervisor.sock"
)

var (
	client *SupervisorClient
	once   sync.Once
)

func fromEnv(addr *string) {
	if eaddr := os.Getenv("SUPERVISOR_ADDR"); eaddr != "" {
		*addr = eaddr
	}
}

func GetClient() *SupervisorClient {
	once.Do(func() {
		client = NewSupervisorClient()
	})
	return client
}

type SupervisorClient struct {
	client proto.SupervisorClient
}

func NewSupervisorClient() *SupervisorClient {
	addr := DefaultSocketAddr
	fromEnv(&addr)
	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		log.Fatalln(err)
	}
	return &SupervisorClient{
		client: proto.NewSupervisorClient(conn),
	}
}

func (c *SupervisorClient) RestartWithContext(ctx context.Context, name string) (err error) {

	if len(name) == 0 {
		name = currentSAImage().RefStr()
	}
	req := &proto.RestartReq{
		Image:    currentSAImage().RefStr(),
		NewImage: name,
	}
	_, err = c.client.Restart(ctx, req)
	return err
}

func (c *SupervisorClient) UpdateWithContext(ctx context.Context, req *proto.UpdateReq) (err error) {
	_, err = c.client.Update(ctx, req)
	return err
}

func (c *SupervisorClient) UpdateSystemWithContext(ctx context.Context, systemImage string) error {
	req := &proto.UpdateSystemReq{
		Image: systemImage,
	}
	_, err := c.client.UpdateSystem(ctx, req)
	return err
}

func (c *SupervisorClient) SystemInfoWithContext(ctx context.Context) (*proto.GetSystemInfoResp, error) {
	return c.client.GetSystemInfo(ctx, &emptypb.Empty{})
}

func (c *SupervisorClient) BackupSmartassistantWithContext(ctx context.Context, backupInfo entity.BackupInfo) (err error) {
	var (
		req proto.BackupReq
	)

	req.Note = backupInfo.Note
	req.BackupPath = backupInfo.BackupPath
	req.FileName = backupInfo.Name
	req.Extensions = strings.Split(backupInfo.Extensions, ",")

	plugins, err := entity.GetInstalledPlugins()
	if err != nil {
		return err
	}

	req.Plugins = []*proto.Plugin{}
	for _, plugin := range plugins {
		req.Plugins = append(req.Plugins, &proto.Plugin{
			ID:      plugin.PluginID,
			Brand:   plugin.Brand,
			Image:   plugin.Image,
			Version: plugin.Version,
		})
	}

	saImage := currentSAImage()
	req.Smartassistant = &proto.Smartassistant{}
	req.Smartassistant.Name = saImage.Name
	req.Smartassistant.Registry = saImage.Registry
	req.Smartassistant.Version = saImage.Tag

	_, err = c.client.Backup(ctx, &req)
	return err
}

func (c *SupervisorClient) RestoreSmartassistantWithContext(ctx context.Context, file string) (err error) {
	var (
		req proto.RestoreReq
	)

	req.File = file

	plugins, err := entity.GetInstalledPlugins()
	if err != nil {
		return err
	}

	req.Plugins = []*proto.Plugin{}
	for _, plugin := range plugins {
		req.Plugins = append(req.Plugins, &proto.Plugin{
			ID:      plugin.PluginID,
			Brand:   plugin.Brand,
			Image:   plugin.Image,
			Version: plugin.Version,
		})
	}

	req.Smartassistant = &proto.Smartassistant{}
	saImage := currentSAImage()
	req.Smartassistant.Name = saImage.Name
	req.Smartassistant.Registry = saImage.Registry
	req.Smartassistant.Version = saImage.Tag

	_, err = c.client.Restore(ctx, &req)
	return err
}

func (c *SupervisorClient) GetExtensionsWithContext(ctx context.Context) (resp *proto.GetExtensionsResp, err error) {
	resp, err = c.client.GetExtensions(ctx, &emptypb.Empty{})
	return
}

func (c *SupervisorClient) EnableRemoteHelpWithContext(ctx context.Context, publicKey []byte) (err error) {
	_, err = c.client.RemoteHelp(ctx, &proto.RemoteHelpReq{Enable: true, PublicKey: publicKey})
	return
}

func (c *SupervisorClient) DisableRemoteHelpWithContext(ctx context.Context) (err error) {
	_, err = c.client.RemoteHelp(ctx, &proto.RemoteHelpReq{Enable: false})
	return
}

func (c *SupervisorClient) RemoteHelpEnabledWithContext(ctx context.Context) (enable bool, err error) {
	var (
		resp *proto.RemoteHelpEnabledResp
	)
	if resp, err = c.client.RemoteHelpEnabled(ctx, &emptypb.Empty{}); err != nil {
		return
	}
	return resp.Enable, err
}
