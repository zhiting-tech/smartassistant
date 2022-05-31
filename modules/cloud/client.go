package cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/datatunnel/v1"
	"github.com/zhiting-tech/smartassistant/pkg/datatunnel/v1/proto"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"google.golang.org/grpc/peer"
)

type ControlStreamClient struct {
	SaID       string
	Key        string
	serviceMap sync.Map
	LogLevel   string
}

func (c *ControlStreamClient) sendAuthAction(stream proto.DatatunnelController_ControlStreamClient, serviceName string) (err error) {

	var (
		bytes []byte
	)
	bytes, err = json.Marshal(&datatunnel.AuthActionData{
		SAID:        c.SaID,
		Key:         c.Key,
		ServiceName: serviceName,
	})
	if err != nil {
		return
	}

	err = stream.Send(&proto.ControlStreamData{
		Action:      datatunnel.AuthAction,
		ActionValue: string(bytes),
	})
	return
}

func (c *ControlStreamClient) HandleStream(stream proto.DatatunnelController_ControlStreamClient) {
	conf := config.GetConf()
	p, ok := peer.FromContext(stream.Context())
	if !ok {
		logger.Warnf("ControlStream can not get addr")
		stream.CloseSend()
		return
	}
	hostname := p.Addr.(*net.TCPAddr).IP
	logger.Debug("ControlStream hostname ", hostname)
	// 遍历需要内网穿透的服务, 进行认证
	if conf.Datatunnel.ExportServices != nil {
		for serviceName, _ := range conf.Datatunnel.ExportServices {
			err := c.sendAuthAction(stream, serviceName)
			if err != nil {
				stream.CloseSend()
				logger.Warning("ControlStream recv err:", err)
				return
			}
		}
	}
	for {
		data, err := stream.Recv()
		if err != nil {
			stream.CloseSend()
			logger.Warning("ControlStream recv err:", err)
			break
		}

		logger.Debugf("Recv action %v, value %v\n", data.Action, data.ActionValue)
		switch data.Action {
		case datatunnel.CreateAction:
			actionData := &datatunnel.CreateActionData{}
			if err = datatunnel.GetActionData(data.ActionValue, actionData); err != nil {
				logger.Debugf("data error %v\n", err)
				break
			}
			c.doCreateClient(fmt.Sprintf("%s:%d", hostname, actionData.Port), actionData.ServiceName)
		case datatunnel.NewAction:
			c.doProxy(data.ActionValue)
		case datatunnel.ErrorAction:
			c.doError(data.ActionValue)
		}
	}
}

func (c *ControlStreamClient) doCreateClient(addr string, serviceName string) {
	conf := config.GetConf()
	serverAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logger.Warnf("ControlStream resolve server addr error %v", err)
		return
	}
	if conf.Datatunnel.ExportServices == nil {
		logger.Warnf("ControlStream export services is nil")
		return
	}
	// 获取对应服务端口
	addr, ok := conf.Datatunnel.GetAddr(serviceName)
	if !ok {
		logger.Warnf("ControlStream can not find service %s\n", serviceName)
		return
	}
	proxyAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logger.Warnf("ControlStream resolve proxy addr error %v", err)
		return
	}
	c.serviceMap.Store(serviceName, datatunnel.NewTunnelClient(&datatunnel.TunnelClientOption{
		ServerAddr: serverAddr,
		ProxyAddr:  proxyAddr,
		LogLevel:   c.LogLevel,
	}))
}

func (c *ControlStreamClient) doProxy(value string) {
	data := datatunnel.NewActionData{}
	err := datatunnel.GetActionData(value, &data)
	if err != nil {
		logger.Warnf("ControlStream GetNewActionData error %v", err)
		return
	}
	val, load := c.serviceMap.Load(data.ServiceName)
	if load {
		client := val.(*datatunnel.TunnelClient)
		// 传入SA连接时发送认证信息的回调函数
		go client.NewProxy(func(c net.Conn) error {
			buf := []byte(data.ConnectionKey)
			n, err := c.Write(buf)
			if err != nil {
				return err
			}

			if n != len(buf) {
				text := fmt.Sprintf("Write error, buf size %d, write count %d", len(buf), n)
				return errors.New(text)
			}

			return nil
		})
	}
}

func (c *ControlStreamClient) doError(text string) {
	logger.Warnf("ControlStream Error Action %v\n", text)
}
