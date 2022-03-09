package role

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"strconv"
	"strings"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"

	"github.com/gin-gonic/gin"
)

// rolePermissionsResp 获取用户权限接口返回数据
type rolePermissionsResp struct {
	Permissions map[string]bool `json:"permissions"`
}

func (resp *rolePermissionsResp) wrap(ps []Permission, up entity.UserPermissions) {
	if len(resp.Permissions) == 0 {
		resp.Permissions = make(map[string]bool)
	}
	for _, v := range ps {
		vp := v.Permission
		strs := []string{vp.Action, vp.Target}
		if vp.Attribute != "" {
			strs = append(strs, vp.Attribute)
		}

		p := strings.Join(strs, "_")
		resp.Permissions[p] = up.IsPermit(vp)
	}
}

// checkSAUpgradePermission 校验sa的固件升级，软件升级权限
func (resp *rolePermissionsResp) checkSAUpgragePermission(up entity.UserPermissions) {
	saDevice, err := entity.GetSaDevice()
	if err != nil {
		return
	}
	ps, err := device.Permissions(saDevice)
	if err != nil {
		return
	}
	permissions := wrapPs(ps)
	for _, permission := range permissions {
		if permission.Permission.Attribute != types.SoftwareUpgrade && permission.Permission.Attribute != types.FwUpgrade {
			continue
		}

		p := fmt.Sprintf("sa_%s", permission.Permission.Attribute)
		resp.Permissions[p] = up.IsPermit(permission.Permission)
	}

}

// UserPermissions 用于处理获取用户权限接口的请求
func UserPermissions(c *gin.Context) {
	var (
		err  error
		resp rolePermissionsResp
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	v := c.Param("id")
	userID, err := strconv.Atoi(v)
	if err != nil {
		return
	}

	if _, err = entity.GetUserByID(userID); err != nil {
		return
	}

	curArea, err := entity.GetAreaByID(session.Get(c).AreaID)
	if err != nil {
		return
	}
	var ps Permissions
	ps, err = getPermissions()
	if err != nil {
		return
	}
	up, err := entity.GetUserPermissions(userID)
	if err != nil {
		logrus.Errorf("wrap err: GetUserPermissions error: %s", err.Error())
		return
	}
	resp.wrap(ps.Device, up)
	for _, v := range ps.DeviceAdvanced.Locations {
		for _, vv := range v.Devices {
			resp.wrap(vv.Permissions, up)
		}
	}

	for _, v := range ps.DeviceAdvanced.Departments {
		for _, vv := range v.Devices {
			resp.wrap(vv.Permissions, up)
		}
	}

	if entity.IsHome(curArea.AreaType) {
		resp.wrap(ps.Area, up)
		resp.wrap(ps.Location, up)
	}else {
		resp.wrap(ps.Company, up)
		resp.wrap(ps.Department, up)
	}
	resp.wrap(ps.Role, up)
	resp.wrap(ps.Scene, up)
	resp.checkSAUpgragePermission(up)

}
