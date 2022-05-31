package datatunnel

import (
	"context"
	"fmt"
	"net"
)

type ProxyClient struct {
	key []byte
}

func NewProxyClient(key []byte) *ProxyClient {
	return &ProxyClient{key: key}
}

func (c *ProxyClient) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if network != "tcp" {
		return nil, fmt.Errorf("unsupported network")
	}

	var d net.Dialer
	conn, err := d.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	if err := c.connect(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func (c *ProxyClient) connect(conn net.Conn) (err error) {
	var (
		p    *proxyConnectProtocol = &proxyConnectProtocol{}
		data []byte
	)

	p.Add(c.key)
	if data, err = p.Encode(); err != nil {
		return
	}
	if _, err := conn.Write(data); err != nil {
		conn.Close()
		return err
	}

	return nil
}
