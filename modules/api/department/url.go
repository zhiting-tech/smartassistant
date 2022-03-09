package department

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/device"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
)

// RegisterDepartmentRouter 注册与部门相关的路由及其处理函数
func RegisterDepartmentRouter(r gin.IRouter) {
	departmentsGroups := r.Group("departments", middleware.RequireAccount)

	departmentGroup := departmentsGroups.Group(":id", requireBelongsToUser)
	departmentGroup.PUT("", middleware.RequirePermission(types.DepartmentUpdate), UpdateDepartment)
	departmentGroup.DELETE("", middleware.RequirePermission(types.DepartmentUpdate), DelDepartment)
	departmentGroup.GET("", middleware.RequirePermission(types.DepartmentGet), InfoDepartment)
	departmentGroup.POST("/users", middleware.RequirePermission(types.DepartmentAddUser), AddDepartmentUser)
	departmentGroup.GET("/devices", device.ListLocationDevices)

	departmentsGroups.GET("", ListDepartment)
	departmentsGroups.POST("", middleware.RequirePermission(types.DepartmentAdd), AddDepartment)
	departmentsGroups.PUT("", middleware.RequirePermission(types.DepartmentUpdateOrder), DepartmentOrder)

}

// requireBelongsToUser 操作部门需要部门属于用户的公司
func requireBelongsToUser(c *gin.Context) {
	user := session.Get(c)

	departmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.BadRequest), nil)
		c.Abort()
		return
	}

	department, err := entity.GetDepartmentByID(departmentID)
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.InternalServerErr), nil)
		c.Abort()
		return
	}
	if department.AreaID == user.AreaID {
		c.Next()
	} else {
		response.HandleResponse(c, errors.New(status.Deny), nil)
		c.Abort()
	}
}
