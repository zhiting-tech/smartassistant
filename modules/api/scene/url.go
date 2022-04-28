// Package scene 设备场景
package scene

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
)

// InitSceneRouter 注册与场景相关的路由及其处理函数
func InitSceneRouter(r gin.IRouter) {
	sceneGroup := r.Group("scenes", middleware.RequireAccount)
	{
		sceneGroup.POST("", CreateScene)
		sceneGroup.DELETE(":id", requireBelongsToUser, DeleteScene)
		sceneGroup.PUT(":id", requireBelongsToUser, middleware.RequirePermission(types.SceneUpdate), UpdateScene)
		sceneGroup.GET("", ListScene)
		sceneGroup.GET(":id", requireBelongsToUser, InfoScene)
		sceneGroup.POST(":id/execute", requireBelongsToUser, ExecuteScene)
		sceneGroup.PUT("", middleware.RequirePermission(types.SceneUpdate), OrderScene)
	}

	r.GET("scene_logs", middleware.RequireAccount, ListSceneTaskLog)
}

// requireBelongsToUser 操作场景需要与用户属于同一个家庭
func requireBelongsToUser(c *gin.Context) {
	u := session.Get(c)
	sceneID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.BadRequest), nil)
		c.Abort()
		return
	}

	scene, err := entity.GetSceneById(sceneID)
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.InternalServerErr), nil)
		c.Abort()
		return
	}

	if u.AreaID != scene.AreaID {
		response.HandleResponse(c, errors.New(status.Deny), nil)
		c.Abort()
	} else {
		c.Next()
	}
}
