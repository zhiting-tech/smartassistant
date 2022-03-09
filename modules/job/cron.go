package job

import (
	"context"
	cron "github.com/robfig/cron/v3"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type Server struct {
}

func NewJobServer() *Server {
	return &Server{}
}

func (s *Server) Run(ctx context.Context) {
	go func() {
		// 实例化Cron定时器
		crontab := cron.New()
		// 添加定时任务, * * * * * 是 crontab,表示每天23点59分执行一次 定时删除超过七天的日志
		crontab.AddFunc("59 23 * * *", LogRemove)
		// 启动定时器
		crontab.Start()
	}()

	<-ctx.Done()
	logger.Warning("job server stopped")
}
