package extension

import (
	pb "github.com/zhiting-tech/smartassistant/pkg/extension/proto"
	"sync"
)

var (
	once  	sync.Once
	extensionServer *ExtensionServer
)

func GetExtensionServer() *ExtensionServer {
	once.Do(func() {
		extensionServer = &ExtensionServer{
			NotifyChans: make(map[chan pb.SAEventInfo]struct{}),
		}
	})
	return extensionServer
}
