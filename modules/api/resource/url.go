package resource

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
)

// RegisterResourceRouter 系统资源监控
func RegisterResourceRouter(r gin.IRouter) {
	resourceGroup := r.Group("resources")
	resourceAuthGroup := resourceGroup.Use(middleware.RequireOwner)
	resourceAuthGroup.GET("", ListResource)
	resourceAuthGroup.PUT("containers/:id", RestartContainer)
}
