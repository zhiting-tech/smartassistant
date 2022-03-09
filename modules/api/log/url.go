package log

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
)

func RegisterLogRouter(r gin.IRouter) {
	logGroup := r.Group("log", middleware.RequireAccount)
	{
		// 日志列表
		logGroup.GET("", LogList)
		// 上传日志
		logGroup.PUT("", middleware.RequireOwner, Upload)
	}
}
