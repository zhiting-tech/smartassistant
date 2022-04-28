package app

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/types"
)

// RegisterAppRouter 用于第三方平台相关的路由及其处理函数
func RegisterAppRouter(r gin.IRouter) {
	areasGroup := r.Group("apps", middleware.RequireAccountWithScope(types.ScopeArea))
	areasGroup.GET("", ListApp)
	areasGroup.DELETE(":id/areas/:area_id", UnbindApp)
}
