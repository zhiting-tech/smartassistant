package logreplay

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type LogHttpServer struct {
	addr      string
	ginEngine *gin.Engine
}

// 日志采集独立服务
func NewLogHttpServer() *LogHttpServer {
	conf := config.GetConf()
	r := gin.Default()
	r.GET("/log", LogReceiver)
	return &LogHttpServer{
		ginEngine: r,
		addr:      conf.SmartAssistant.LogHttpAddress(),
	}
}

func (s *LogHttpServer) LogSrcRun(ctx context.Context) {
	logger.Infof("starting log http server on %s", s.addr)
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.ginEngine,
	}

	// 启动log http服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logger.Infof("log http listen: %s\n", err)
		}
	}()
	<-ctx.Done()

	stop, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(stop)
	logger.Warning("log http server stopped")
}
