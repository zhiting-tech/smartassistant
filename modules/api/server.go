package api

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/logreplay"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils/url"
	"github.com/zhiting-tech/smartassistant/modules/websocket"
	"github.com/zhiting-tech/smartassistant/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type HttpServer struct {
	addr      string
	ginEngine *gin.Engine
}

func NewHttpServer(ws gin.HandlerFunc) *HttpServer {
	conf := config.GetConf()
	r := gin.Default()

	versionParamKey := "_version"

	apiGroup := r.Group("api")
	apiGroup.Use(
		otelgin.Middleware("api"),
		// 记录请求日志
		middleware.AccessLog(),
		middleware.VersionMiddleware(versionParamKey, types.Version, types.MinVersion),
	)
	apiVersionGroup := apiGroup.Group("/:" + versionParamKey)
	loadModules(apiVersionGroup)
	loadModules(apiGroup)

	// 注册websocket命令
	websocket.RegisterCmd()
	websocketGroup := r.Group("/ws")
	websocketGroup.Use(otelgin.Middleware("websocket"), middleware.RequireToken)
	websocketGroup.GET("", ws)

	r.GET("/log", logreplay.LogReceiver)

	// 静态文件
	r.Static(url.BackendStaticPath(), "./static")
	r.Static(url.FilePath(),
		filepath.Join(
			config.GetConf().SmartAssistant.RuntimePath,
			"run", "smartassistant", "file"),
	)

	// 代理插件
	r.Any("/plugin/:plugin/*path", middleware.ProxyToPlugin)

	return &HttpServer{
		ginEngine: r,
		addr:      conf.SmartAssistant.HttpAddress(),
	}
}

func (s *HttpServer) Run(ctx context.Context) {
	logger.Infof("starting http server on %s", s.addr)
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.ginEngine,
	}

	// 启动http服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logger.Infof("http listen: %s\n", err)
		}
	}()
	<-ctx.Done()

	stop, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(stop)
	logger.Warning("http server stopped")
}
