package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/zhiting-tech/smartassistant/modules/websocket"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type HttpServer struct {
	addr      string
	ginEngine *gin.Engine
}

func NewHttpServer(ws gin.HandlerFunc) *HttpServer {
	conf := config.GetConf()
	r := gin.Default()

	// 记录请求日志
	r.Use(otelgin.Middleware("api"), middleware.AccessLog())
	apiGroup := r.Group("api")
	// 注册websocket命令
	websocket.RegisterCmd()
	r.GET("/ws", middleware.RequireToken, ws)
	loadModules(apiGroup)
	apiGroup.Static(fmt.Sprintf("static/%s/sa", conf.SmartAssistant.ID), "./static")

	// 代理插件
	apiGroup.Any("/plugin/:plugin/*path", middleware.ProxyToPlugin)

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
