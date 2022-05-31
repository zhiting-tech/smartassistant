package cloud

import (
	"context"
	"net"
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/datatunnel/v2"
	"github.com/zhiting-tech/smartassistant/pkg/datatunnel/v2/control"
	"github.com/zhiting-tech/smartassistant/pkg/datatunnel/v2/proto"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

func RunProxyClient(ctx context.Context) {
	client := datatunnel.NewProxyControlClient(
		datatunnel.WithProxyAddr(config.GetConf().Datatunnel.ProxyManagerAddr),
		datatunnel.WithLogger(logger.NewEntry()),
		datatunnel.WithClientHook(datatunnel.ProxyControlClientHook{
			OnConnected: func(c *datatunnel.ProxyControlClient, pcsc *control.ProxyControlStreamContext) error {
				if len(config.GetConf().SmartAssistant.ID) <= 0 || len(config.GetConf().SmartAssistant.Key) <= 0 {
					return control.NewControlError(datatunnel.InvalidSAIDOrSAKey)
				}

				return c.Authenticate(pcsc, &proto.AuthenticateRequest{
					SAID:  config.GetConf().SmartAssistant.ID,
					SAKey: config.GetConf().SmartAssistant.Key,
				})
			},
		}),
	)

	for serviceName, _ := range config.GetConf().Datatunnel.ExportServices {
		addr, ok := config.GetConf().Datatunnel.GetAddr(serviceName)
		if !ok {
			continue
		}

		host, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			continue
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		client.RegisterProxyServices(datatunnel.ProxyService{
			ServiceName: serviceName,
			ServicePort: port,
			ServiceHost: host,
		})
	}

	client.Run(ctx, config.GetConf().Datatunnel.ControlServerAddr)
}
