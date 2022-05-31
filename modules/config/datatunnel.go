package config

import (
	"fmt"
	"net"
	"strconv"

	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type Datatunnel struct {
	ControlServerAddr string            `json:"control_server_addr" yaml:"control_server_addr"`
	ProxyManagerAddr  string            `json:"proxy_manager_addr" yaml:"proxy_manager_addr"`
	ExportServices    map[string]string `json:"export_services" yaml:"export_services"`
}

func (t *Datatunnel) GetAddr(serviceName string) (addr string, ok bool) {
	var (
		val string
	)
	val, ok = t.ExportServices[serviceName]
	if !ok {
		return
	}

	port, err := strconv.ParseInt(val, 10, 32)
	if err == nil {
		addr = fmt.Sprintf("127.0.0.1:%d", port)
		return
	}

	addr = val
	return
}

func (t *Datatunnel) GetPort(serviceName string) (port int, ok bool) {
	val, ok := t.GetAddr(serviceName)
	if !ok {
		return
	}
	addr, err := net.ResolveTCPAddr("tcp", val)
	if err != nil {
		logger.Error("resolve tcp err: ", err)
		return
	}

	port = addr.Port
	return port, true
}
