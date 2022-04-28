// Package device 设备，包括SA状态
package device

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

// RegisterDeviceRouter 注册与设备相关的路由及其处理函数
func RegisterDeviceRouter(r gin.IRouter) {

	deviceAuthGroup := r.Group("devices", middleware.RequireAccountWithScope(types.ScopeDevice))
	deviceAuthGroup.GET("", ListAllDevice)
	deviceAuthGroup.PUT(":id", requireBelongsToUser, UpdateDevice)
	deviceAuthGroup.GET(":id", requireBelongsToUser, InfoDevice)
	deviceAuthGroup.GET(":id/logo", requireBelongsToUser, InfoDeviceLogo)

	// 设备型号列表（按分类分组）
	r.GET("device/types/major", MajorTypeList)
	r.GET("device/types/minor", MinorTypeList)

	// 检查SA是否已绑定
	r.GET("/check", CheckSaDevice)
	// 检查SA是否对应的云端同步下来的家庭
	r.POST("/check", IsAccessAllow)

	// 添加SA
	r.POST("/sa", AddDevice)

}

// requireBelongsToUser 操作的设备需要设备属于用户的家庭
func requireBelongsToUser(c *gin.Context) {
	user, err := entity.GetUserByID(session.Get(c).UserID)
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.InternalServerErr), nil)
		c.Abort()
		return
	}

	deviceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.BadRequest), nil)
		c.Abort()
		return
	}

	device, err := entity.GetDeviceByID(deviceID)
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.InternalServerErr), nil)
		c.Abort()
		return
	}
	if device.AreaID == user.AreaID {
		c.Next()
	} else {
		response.HandleResponse(c, errors.New(status.Deny), nil)
		c.Abort()
	}
}
