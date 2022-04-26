// Package area 公司/家庭
package area

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// RegisterAreaRouter 用于注册与家庭相关的路由及其处理函数
func RegisterAreaRouter(r gin.IRouter) {
	areasGroup := r.Group("areas", middleware.RequireAccountWithScope(types.ScopeArea))
	areasGroup.GET("", ListArea)

	areaGroup := areasGroup.Group(":id", requireBelongsToUser)
	areaGroup.PUT("", middleware.RequirePermission(types.AreaUpdateName, types.AreaUpdateCompanyName), UpdateArea)
	areaGroup.DELETE("", middleware.RequireOwner, DelArea)
	areaGroup.GET("", InfoArea)
	areaGroup.DELETE("/users/:user_id", QuitArea)
	areaGroup.DELETE("/device/sa", AreaDelSa)

	r.POST("/sync", middleware.RequireOwner, DataSync)
}

// requireBelongsToUser 操作家庭需要当前用户属于该家庭
func requireBelongsToUser(c *gin.Context) {
	user, err := entity.GetUserByID(session.Get(c).UserID)
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.InternalServerErr), nil)
		c.Abort()
		return
	}
	areaID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.BadRequest), nil)
		c.Abort()
		return
	}
	if user.BelongsToArea(areaID) {
		c.Next()
	} else {
		response.HandleResponse(c, errors.New(status.Deny), nil)
		c.Abort()
	}
}
