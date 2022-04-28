package file

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
)

// RegisterFileRouter 注册与文件相关的路由及其处理函数
func RegisterFileRouter(r gin.IRouter) {
	r.POST("files",  middleware.RequireAccount, FileUpload)
}
