// Package reverseproxy 反向代理
package reverseproxy

import (
	"errors"
	"fmt"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

var (
	ErrorUpstreamNotValid = errors.New("upstream not valid")
	ErrorUpstreamExists   = errors.New("upstream exists")
	ErrorUpstreamNotFound = errors.New("upstream not found")
)

var m *Manager
var managerOnce sync.Once

type Manager struct {
	servers map[string]Upstream
	mu      sync.RWMutex
}

type Upstream struct {
	Proxy *httputil.ReverseProxy
}

func GetManager() *Manager {
	managerOnce.Do(func() {
		m = &Manager{
			servers: make(map[string]Upstream),
		}
	})
	return m
}

func (m *Manager) GetUpstream(path string) (up Upstream, err error) {

	logger.Debugf("get upstream %s", path)
	m.mu.RLock()
	defer m.mu.RUnlock()
	up, ok := m.servers[path]
	if !ok {
		err = ErrorUpstreamNotFound
		logger.Errorf("upstream %s not found", path)
		return
	}
	logger.Debugf("upstream %s found", path)
	return
}

// RegisterUpstream 注册一个反向代理后端
// serverUrl 支持 :8090, 127.0.0.1:8090，http://127.0.0.1:8090 格式
func (m *Manager) RegisterUpstream(path string, serverUrl string) (err error) {
	if len(path) == 0 || len(serverUrl) == 0 {
		err = ErrorUpstreamNotValid
		return
	}
	if strings.HasPrefix(serverUrl, ":") {
		// 只提供端口
		serverUrl = fmt.Sprintf("http://127.0.0.1%s", serverUrl)
	} else if !strings.HasPrefix(serverUrl, "http") {
		serverUrl = fmt.Sprintf("http://%s", serverUrl)
	}
	su, err := url.Parse(serverUrl)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.servers[path]
	if ok {
		err = ErrorUpstreamExists
		return
	}
	up := Upstream{
		Proxy: httputil.NewSingleHostReverseProxy(su),
	}
	m.servers[path] = up
	return
}

func (m *Manager) UnregisterUpstream(path string) (err error) {
	if len(path) == 0 {
		err = ErrorUpstreamNotValid
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.servers, path)
	return
}

func RegisterUpstream(path string, serverUrl string) (err error) {
	return GetManager().RegisterUpstream(path, serverUrl)
}

func UnregisterUpstream(path string) (err error) {
	return GetManager().UnregisterUpstream(path)
}
