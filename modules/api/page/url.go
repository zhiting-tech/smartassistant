// Package page 公共页面
package page

import (
	"github.com/gin-gonic/gin"
)

func RegisterPageRouter(r gin.IRouter) {
	pageGroup := r.Group("/pages")
	{
		// TODO:判断当前用户是否有创建场景的权限
		pageGroup.GET("create_scene", CreateSceneGuideRedirect)
	}
}
