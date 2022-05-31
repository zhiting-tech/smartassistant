package supervisor

import (
	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

func RegisterSupervisorRouter(r gin.IRouter) {
	supervisorGroup := r.Group("supervisor", middleware.RequireOwner)
	{
		supervisorGroup.GET("backups", ListBackup)
		supervisorGroup.POST("backups", AddBackup)
		supervisorGroup.DELETE("backups", DeleteBackup)
		supervisorGroup.POST("backups/restore", Restore)
		supervisorGroup.GET("backups/paths", ListBackupPath)
		supervisorGroup.POST("backups/mount", MountDisk)
	}
	r.GET("supervisor/update", middleware.RequireAccount, middleware.RequirePermission(getSwUpgradePermission()), UpdateInfo)
	r.GET("supervisor/update/latest", middleware.RequireAccount, middleware.RequirePermission(getSwUpgradePermission()), UpdateLastVersion)
	r.POST("supervisor/update", middleware.RequireAccount, middleware.RequirePermission(getSwUpgradePermission()), Update)

	r.POST("supervisor/firmware/update", middleware.RequireAccount, middleware.RequirePermission(getFwUpgradePermission()), UpdateSystem)
	r.GET("supervisor/firmware/update", middleware.RequireAccount, middleware.RequirePermission(getFwUpgradePermission()), GetSystemInfo)
	r.GET("supervisor/firmware/update/latest", middleware.RequireAccount, middleware.RequirePermission(getFwUpgradePermission()), GetSystemLastVersion)
}

// getSwUpgradePermission 获取软件升级权限
func getSwUpgradePermission() (p types.Permission) {
	device, err := entity.GetSaDevice()
	if err != nil {
		logger.Error(err)
		return
	}

	return types.NewDeviceSoftwareUpgrade(device.ID)
}

func getFwUpgradePermission() (p types.Permission) {
	device, err := entity.GetSaDevice()
	if err != nil {
		logger.Error(err)
		return
	}

	return types.NewDeviceFwUpgrade(device.ID)
}
