package cloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/datatunnel/v1/proto"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	DefaultGrpcIdleTime       = 10 * time.Second
	DefaultGrpcPingAckTimeout = 20 * time.Second
)

func StartDataTunnel(ctx context.Context) {
	var level string
	conf := config.GetConf()

	hostname := conf.SmartCloud.Domain
	index := strings.Index(hostname, ":")
	if index > 0 {
		hostname = hostname[:index]
	}

	sleepTime := 2
	target := fmt.Sprintf("%s:%d", hostname, conf.SmartCloud.GRPCPort)
	// 空闲时间10秒则发送ping,发送ping后20秒内收不到ack视为断开连接
	kacp := keepalive.ClientParameters{
		Time:                DefaultGrpcIdleTime,
		Timeout:             DefaultGrpcPingAckTimeout,
		PermitWithoutStream: true,
	}
	conn, err := grpc.Dial(target, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		logger.Warning("grpc connect err:", err)
		return
	}

	if conf.Debug {
		level = "debug"
	}
	sleepTime = 2
	grpcClient := proto.NewDatatunnelControllerClient(conn)
	streamClient := &ControlStreamClient{
		SaID:     conf.SmartAssistant.ID,
		Key:      conf.SmartAssistant.Key,
		LogLevel: level,
	}
	for {

		stream, err := grpcClient.ControlStream(ctx)
		if err != nil {
			logger.Warning("ControlStream err:", err)
			sleepTime = sleepTime * 2
			if sleepTime > 600 {
				sleepTime = 600
			}
		} else {
			sleepTime = 2
		}

		if stream != nil {
			streamClient.HandleStream(stream)
		}

		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}
