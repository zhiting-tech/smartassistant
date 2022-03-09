package supervisor

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/supervisor/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	return &SupervisorClient{
		client: proto.NewSupervisorClient(conn),
	}
}

func (c *SupervisorClient) Restart(name string) (err error) {

	if len(name) == 0 {
		name = currentSAImage().RefStr()
	}
	req := &proto.RestartReq{
		Image:    currentSAImage().RefStr(),
		NewImage: name,
	}
	_, err = c.client.Restart(context.Background(), req)
	return err
}

func (c *SupervisorClient) Update(req *proto.UpdateReq) (err error) {
	_, err = c.client.Update(context.Background(), req)
	return err
}

func (c *SupervisorClient) UpdateSystem(systemImage string) error {
	req := &proto.UpdateSystemReq{
		Image: systemImage,
	}
	_, err := c.client.UpdateSystem(context.TODO(), req)
	return err
}

func (c *SupervisorClient) SystemInfo() (*proto.GetSystemInfoResp, error) {
	return c.client.GetSystemInfo(context.TODO(), &emptypb.Empty{})
}

func (c *SupervisorClient) BackupSmartassistant(note string) (err error) {
	var (
		req proto.BackupReq
	)

	req.Note = note

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

	_, err = c.client.Backup(context.TODO(), &req)
	return err
}

func (c *SupervisorClient) RestoreSmartassistant(file string) (err error) {
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

	_, err = c.client.Restore(context.TODO(), &req)
	return err
}

func (c *SupervisorClient) GetExtensions() (resp *proto.GetExtensionsResp, err error) {
	resp, err = c.client.GetExtensions(context.TODO(), &emptypb.Empty{})
	return
}
