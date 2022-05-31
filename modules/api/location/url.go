// Package location 房间
package location

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/device"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// RegisterLocationRouter 注册与房间相关的路由及其处理函数
func RegisterLocationRouter(r gin.IRouter) {
	locationsGroup := r.Group("locations", middleware.RequireAccount)

	locationGroup := locationsGroup.Group(":id", requireBelongsToUser)
	locationGroup.PUT("", middleware.RequirePermission(types.LocationUpdateName), UpdateLocation)
	locationGroup.DELETE("", middleware.RequirePermission(types.LocationDel), DelLocation)
	locationGroup.GET("", middleware.RequirePermission(types.LocationGet), InfoLocation)
	locationGroup.GET("/devices", device.ListLocationDevices)
	locationGroup.PUT("/devices", middleware.RequirePermission(types.DeviceUpdateOrder), updateLocationDevices)

	locationsGroup.GET("", ListLocation)
	locationsGroup.POST("", middleware.RequirePermission(types.LocationAdd), AddLocation)
	locationsGroup.PUT("", middleware.RequirePermission(types.LocationUpdateOrder), LocationOrder)

	r.GET("location_tmpl", ListDefaultLocation)
}

// requireBelongsToUser 操作房间需要房间属于用户的家庭
func requireBelongsToUser(c *gin.Context) {
	user := session.Get(c)

	locationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.BadRequest), nil)
		c.Abort()
		return
	}

	location, err := entity.GetLocationByID(locationID)
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.InternalServerErr), nil)
		c.Abort()
		return
	}
	if location.AreaID == user.AreaID {
		c.Next()
	} else {
		response.HandleResponse(c, errors.New(status.Deny), nil)
		c.Abort()
	}
}
