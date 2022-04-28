package addr

import (
	"errors"
	"net"
)

// LocalIP 获取本地地址 TODO 暂时提供给插件获取本地地址，后续取消host模式后修改
func LocalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, i := range ifaces {
		if i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() == nil {
					continue
				}
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("no ip found")
}
