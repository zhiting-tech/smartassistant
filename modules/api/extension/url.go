package extension

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
)

func RegisterExtensionRouter(r gin.IRouter) {
	extensionGroup := r.Group("extensions", middleware.RequireAccount)
	extensionGroup.GET("", ListExtension)
}
