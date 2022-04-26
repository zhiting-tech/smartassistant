package main

import (
	"context"
	"flag"
	"github.com/zhiting-tech/smartassistant/modules/job"
	"github.com/zhiting-tech/smartassistant/modules/utils/backup"
	"github.com/zhiting-tech/smartassistant/pkg/filebrowser"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhiting-tech/smartassistant/pkg/trace"

	"github.com/sirupsen/logrus"

	"github.com/zhiting-tech/smartassistant/modules/api"
	"github.com/zhiting-tech/smartassistant/modules/api/setting"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/event"
	"github.com/zhiting-tech/smartassistant/modules/extension"
	"github.com/zhiting-tech/smartassistant/modules/logreplay"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/sadiscover"
	"github.com/zhiting-tech/smartassistant/modules/task"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/websocket"
	"github.com/zhiting-tech/smartassistant/pkg/analytics"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/reverseproxy"
)

var configFile = flag.String("c", "/mnt/data/zt-smartassistant/config/smartassistant.yaml", "config file")

func main() {
	flag.Parse()
	conf := config.InitConfig(*configFile)
	ctx, cancel := context.WithCancel(context.Background())
	initLog(conf.Debug)
	trace.Init("smartassistant")
	logger.Infof("starting smartassistant %v", types.Version)
	// 优先使用单例模式，循环引用通过依赖注入解耦
	taskManager := task.GetManager()
	wsServer := websocket.NewWebSocketServer()
	httpServer := api.NewHttpServer(wsServer.AcceptWebSocket)
	saDiscoverServer := sadiscover.NewSaDiscoverServer()
	filebrowser.GetFBOrInit()

	// 启动日志转发功能
	logreplay.GetLogPlayer().EnableSave()
	go logreplay.GetLogPlayer().Run(ctx)
	// 启动定时任务和队列任务
	go job.GetJobServer().Run(ctx)
	// 启动数据埋点服务
	go analytics.Start(conf)

	go wsServer.Run(ctx)
	// 新建插件manager并设为全局
	pluginManager := plugin.NewManager()
	plugin.SetGlobalManager(pluginManager)
	// 新建插件client并设为全局
	pluginClient := plugin.NewClient()
	plugin.SetGlobalClient(pluginClient)

	backup.CheckBackupInfo()
	// 新建服务发现
	discovery := plugin.NewDiscovery(pluginClient)
	go discovery.Listen(ctx)

	go httpServer.Run(ctx)
	go saDiscoverServer.Run(ctx)
	go extension.GetExtensionServer().Run(ctx)
	event.RegisterEventFunc(wsServer)

	// 等待其他服务启动完成
	time.Sleep(3 * time.Second)

	go taskManager.Run(ctx)

	reverseproxy.RegisterUpstream(types.CloudDisk, types.CloudDiskAddr)
	// 如果已配置，则尝试连接 SmartCloud
	if len(conf.SmartCloud.Domain) > 0 {
		// 尝试发送认证token给SC
		go setting.SendAreaAuthToSC()
	}

	if len(conf.SmartCloud.Domain) > 0 && conf.SmartCloud.GRPCPort > 0 && len(conf.SmartAssistant.ID) > 0 {
		// 启动数据通道
		go cloud.StartDataTunnel(ctx)
	}

	logger.Info("SmartAssistant started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sig:
		// Exit by user
	}
	logger.Info("shutting down.")
	cancel()
	time.Sleep(3 * time.Second)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func initLog(debug bool) {
	fields := logrus.Fields{
		"app":   "smartassistant",
		"sa_id": config.GetConf().SmartAssistant.ID,
	}
	if debug {
		logger.InitLogger(os.Stderr, logrus.DebugLevel, fields, debug)
	} else {
		logger.InitLogger(os.Stderr, logrus.InfoLevel, fields, debug)
	}
}
