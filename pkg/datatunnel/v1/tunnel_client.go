package datatunnel

import (
	"net"

	"github.com/sirupsen/logrus"
)

type TunnelClient struct {
	serverAddr   *net.TCPAddr
	proxyAddr    *net.TCPAddr
	retry        int
	logger       *logrus.Logger
	lingerSecond int // tcp参数SO_LINGER
}

type TunnelClientOption struct {
	// 服务端和需要代理的地址
	ServerAddr *net.TCPAddr
	ProxyAddr  *net.TCPAddr

	// 重试次数
	Retry int

	LogLevel string

	LingerSecond int // tcp参数SO_LINGER,单位秒
}

func NewTunnelClient(option *TunnelClientOption) *TunnelClient {
	logger := logrus.New()
	logger.SetLevel(getLogrusLevel(option.LogLevel))
	if option.LingerSecond == 0 {
		option.LingerSecond = DefaultLingerSecond
	}
	return &TunnelClient{
		serverAddr:   option.ServerAddr,
		proxyAddr:    option.ProxyAddr,
		retry:        option.Retry,
		logger:       logger,
		lingerSecond: option.LingerSecond,
	}
}

func (c *TunnelClient) NewProxy(sendKey func(net.Conn) error) {
	done := make(chan interface{}, 2)
	var conn1, conn2 *net.TCPConn
	go func() {
		var err error
		c.logger.Debug("dial server addr ", c.serverAddr)
		conn1, err = net.DialTCP("tcp", nil, c.serverAddr)
		if err != nil {
			c.logger.Warnf("New proxy dial error %v\n", c.serverAddr)
		} else {
			if c.lingerSecond > 0 {
				conn1.SetLinger(c.lingerSecond)
			}
			conn1.SetKeepAlive(true)
			conn1.SetKeepAlivePeriod(DefaultTCPKeepAlivePeriod)
			if sendKey != nil {
				err = sendKey(conn1)
				if err != nil {
					c.logger.Warnf("Send Key error %v\n", err)
					conn1.Close()
					conn1 = nil
				}
			}
		}
		done <- nil
	}()
	go func() {
		var err error
		c.logger.Debug("dial local addr ", c.proxyAddr)
		conn2, err = net.DialTCP("tcp", nil, c.proxyAddr)
		if err != nil {
			c.logger.Warnf("New proxy dial error %v\n", c.proxyAddr)
		} else {
			if c.lingerSecond > 0 {
				conn2.SetLinger(c.lingerSecond)
			}
			conn2.SetKeepAlive(true)
			conn2.SetKeepAlivePeriod(DefaultTCPKeepAlivePeriod)
		}
		done <- nil
	}()

	<-done
	<-done
	close(done)

	if conn1 == nil || conn2 == nil {
		if conn1 != nil {
			conn1.Close()
		}
		if conn2 != nil {
			conn2.Close()
		}

		c.logger.Debugln("New proxy error")
		return
	}

	Proxy(conn1, conn2)
}
