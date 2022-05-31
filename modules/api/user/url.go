// Package user 用户管理，权限，邀请别人假如
package user

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/api/role"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// RegisterUserRouter 注册与用户相关的路由及其处理函数
func RegisterUserRouter(r gin.IRouter) {
	usersGroup := r.Group("/users", middleware.RequireAccountWithScope(types.ScopeUser))
	usersGroup.GET("", ListUser)

	userGroup := usersGroup.Group(":id", requireSameArea)
	userGroup.GET("", InfoUser)
	userGroup.PUT("", UpdateUser)
	userGroup.DELETE("", middleware.RequirePermission(types.AreaDelMember), DelUser)
	userGroup.POST("/invitation/code", middleware.RequirePermission(types.AreaGetCode), GetInvitationCode)
	userGroup.GET("/permissions", role.UserPermissions)
	userGroup.PUT("/owner", TransferOwner)
	userGroup.PUT("/departments", UserInfoDepartment)
	userGroup.GET("/common_devices", getUserCommonDevices)
	userGroup.PUT("/common_devices", updateUserCommonDevices)

	invitationGroup := r.Group("/invitation")
	{
		invitationGroup.POST("/check", CheckQrCode)
	}
	// 验证码接口，用于第三方云绑定的校验
	r.POST("verification/code", middleware.RequireAccount, GetVerificationCode)
}

// requireInSameArea 请求用户api需要在同一个家庭下
func requireSameArea(c *gin.Context) {

	u := session.Get(c)

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.BadRequest), nil)
		c.Abort()
		return
	}

	user, err := entity.GetUserByID(userID)
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.InternalServerErr), nil)
		c.Abort()
		return
	}

	if u.AreaID != user.AreaID {
		response.HandleResponse(c, errors.New(status.Deny), nil)
		c.Abort()
	} else {
		c.Next()
	}

}
