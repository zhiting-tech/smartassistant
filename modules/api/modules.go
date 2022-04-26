package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/app"
	"github.com/zhiting-tech/smartassistant/modules/api/area"
	"github.com/zhiting-tech/smartassistant/modules/api/auth"
	"github.com/zhiting-tech/smartassistant/modules/api/brand"
	"github.com/zhiting-tech/smartassistant/modules/api/cloud"
	"github.com/zhiting-tech/smartassistant/modules/api/department"
	"github.com/zhiting-tech/smartassistant/modules/api/device"
	"github.com/zhiting-tech/smartassistant/modules/api/extension"
	"github.com/zhiting-tech/smartassistant/modules/api/file"
	"github.com/zhiting-tech/smartassistant/modules/api/location"
	"github.com/zhiting-tech/smartassistant/modules/api/log"
	"github.com/zhiting-tech/smartassistant/modules/api/page"
	"github.com/zhiting-tech/smartassistant/modules/api/plugin"
	"github.com/zhiting-tech/smartassistant/modules/api/resource"
	"github.com/zhiting-tech/smartassistant/modules/api/role"
	"github.com/zhiting-tech/smartassistant/modules/api/scene"
	"github.com/zhiting-tech/smartassistant/modules/api/scope"
	"github.com/zhiting-tech/smartassistant/modules/api/setting"
	"github.com/zhiting-tech/smartassistant/modules/api/smartcloud"
	"github.com/zhiting-tech/smartassistant/modules/api/supervisor"
	"github.com/zhiting-tech/smartassistant/modules/api/user"
)

// loadModules 注册路由及其处理函数
func loadModules(r gin.IRouter) {
	location.RegisterLocationRouter(r)
	brand.RegisterBrandRouter(r)
	device.RegisterDeviceRouter(r)
	area.RegisterAreaRouter(r)
	user.RegisterUserRouter(r)
	scope.RegisterScopeRouter(r)
	role.RegisterRoleRouter(r)
	scene.InitSceneRouter(r)
	page.RegisterPageRouter(r)
	cloud.InitCloudRouter(r)
	setting.RegisterSettingRouter(r)
	supervisor.RegisterSupervisorRouter(r)
	auth.InitAuthRouter(r)
	plugin.RegisterPluginRouter(r)
	smartcloud.InitSmartCloudRouter(r)
	log.RegisterLogRouter(r)
	department.RegisterDepartmentRouter(r)
	extension.RegisterExtensionRouter(r)
	file.RegisterFileRouter(r)
	app.RegisterAppRouter(r)
	resource.RegisterResourceRouter(r)
}
