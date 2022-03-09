package role

import (
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

var (
	RoleNameSizeMax = 20
)

// roleInfo 修改/添加角色接口请求参数
type roleInfo struct {
	entity.RoleInfo
	Permissions *Permissions `json:"permissions,omitempty"`
	IsManager   bool         `json:"is_manager"`
}

// Permissions 角色权限信息
type Permissions struct {
	Device         []Permission   `json:"device"`          // 设备权限设置
	DeviceAdvanced DeviceAdvanced `json:"device_advanced"` // 设备高级权限设置
	Area           []Permission   `json:"area"`            // 家庭权限设置
	Location       []Permission   `json:"location"`        // 区域权限设置
	Role           []Permission   `json:"role"`            // 角色权限设置
	Scene          []Permission   `json:"scene"`           // 场景权限设置
	Company        []Permission   `json:"company"`         // 公司权限设置
	Department     []Permission   `json:"department"`      // 部门权限设置
}

// DeviceAdvanced 设备高级权限信息
type DeviceAdvanced struct {
	Locations []Location `json:"locations"`
	Departments []Location `json:"departments"`
}

// Location 房间信息
type Location struct {
	Name    string   `json:"name"`
	Devices []Device `json:"devices"`
}

// Device 设备信息
type Device struct {
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

// Permission 权限信息
type Permission struct {
	Permission types.Permission `json:"permission"`
	Allow      bool             `json:"allow"` // 是否允许
}

// roleUpdateResp 修改/添加角色接口返回数据
type roleUpdateResp struct {
}

// roleUpdate 用于处理修改/添加角色接口的请求
func roleUpdate(c *gin.Context) {
	var (
		err  error
		req  roleInfo
		resp roleUpdateResp
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if err = c.BindUri(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if err = req.Validate(session.Get(c).AreaID); err != nil {
		return
	}

	var r entity.Role
	if c.Request.Method == http.MethodPut && req.ID != 0 {
		if _, err = entity.GetRoleByID(req.ID); err != nil {
			return
		}
		r, err = entity.UpdateRole(req.ID, req.Name)
	} else {
		// 为某一个家庭添加角色
		sessionUser := session.Get(c)
		r, err = entity.AddRole(req.Name, sessionUser.AreaID)
	}
	if err != nil {
		return
	}
	if req.Permissions == nil {
		return
	}

	for _, v := range req.Permissions.DeviceAdvanced.Locations {
		for _, vv := range v.Devices {
			updatePermission(r, vv.Permissions)
		}
	}

	for _, v := range req.Permissions.DeviceAdvanced.Departments {
		for _, vv := range v.Devices {
			updatePermission(r, vv.Permissions)
		}
	}

	updatePermission(r, req.Permissions.Device)
	updatePermission(r, req.Permissions.Area)
	updatePermission(r, req.Permissions.Location)
	updatePermission(r, req.Permissions.Role)
	updatePermission(r, req.Permissions.Scene)
	updatePermission(r, req.Permissions.Company)
	updatePermission(r, req.Permissions.Department)
}

func updatePermission(role entity.Role, ps []Permission) {
	for _, v := range ps {
		if v.Allow {
			role.AddPermissions(v.Permission)
		} else {
			role.DelPermission(v.Permission)
		}
	}
}

// 参数验证
func (req *roleInfo) Validate(areaID uint64) (err error) {
	// 角色名称必须填写
	if req.Name == "" {
		err = errors.Wrap(err, status.RoleNameInputNilErr)
		return
	}
	//	角色名称长度不能大于20位
	if utf8.RuneCountInString(req.Name) > RoleNameSizeMax {
		err = errors.Wrap(err, status.RoleNameLengthLimit)
		return
	}

	// 角色名称是否重复
	if entity.IsRoleNameExist(req.Name, req.ID, areaID) {
		err = errors.Wrap(err, status.RoleNameExist)
		return
	}

	return
}
