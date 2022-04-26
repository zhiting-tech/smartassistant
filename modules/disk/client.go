package disk

import (
	"context"
	"github.com/zhiting-tech/smartassistant/modules/disk/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	diskManagerAddr = "disk-manager:9090"
)

type DiskManagerClient struct {
	client proto.DiskManagerClient
}

func NewDiskManagerClient() (client *DiskManagerClient, err error) {
	conn, err := grpc.Dial(
		diskManagerAddr,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		return
	}
	return &DiskManagerClient{
		client: proto.NewDiskManagerClient(conn),
	}, nil
}

func (c *DiskManagerClient) GetPhysicalVolumeListWithContext(ctx context.Context) (resp *proto.PhysicalVolumeListResp, err error) {
	return c.client.PhysicalVolumeList(ctx, &proto.Empty{})
}

func (c *DiskManagerClient) MountPhysicalWithContext(ctx context.Context, pvName string) (err error) {
	var (
		req proto.MountPhysicalReq
	)
	req.PVName = pvName
	_, err = c.client.MountPhysical(ctx, &req)
	return
}

func (c *DiskManagerClient) GetPhysicalMountedListWithContext(ctx context.Context) (resp *proto.PhysicalVolumeListResp, err error) {
	return c.client.PhysicalMountedList(ctx, &proto.Empty{})
}
