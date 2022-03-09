package smartcloud

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
)

func InitSmartCloudRouter(r gin.IRouter) {
	scGroup := r.Group("sc", middleware.ValidateSCReq)
	// 用于sc请求找回用户凭证的接口
	scGroup.GET("users/:id/token/:type", GetToken)
}
