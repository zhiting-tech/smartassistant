package department

import (
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/location"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/file"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type infoDepartment struct {
	Users   []departmentUser  `json:"users"`
	Name    string            `json:"name"`
	Devices []location.Device `json:"devices"`
}

type departmentUser struct {
	entity.UserInfo
	IsManager bool `json:"is_manager"` // 是否是主管
}

// InfoDepartment 返回部门信息
func InfoDepartment(c *gin.Context) {
	var (
		err          error
		departmentId int
		infoDevices  []location.Device
		department   entity.Department
		resp         infoDepartment
	)

	defer func() {
		if resp.Devices == nil {
			resp.Devices = make([]location.Device, 0)
		}
		if resp.Users == nil {
			resp.Users = make([]departmentUser, 0)
		}
		response.HandleResponse(c, err, &resp)
	}()

	departmentId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if department, err = entity.GetDepartmentByID(departmentId); err != nil {
		return
	}
	infoUsers, err := GetDepartmentUsers(c, departmentId, session.Get(c).AreaID, department.ManagerID)
	if err != nil {
		return
	}
	if infoDevices, err = GetDeviceByDepartmentID(departmentId, c); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp.Users = infoUsers
	resp.Name = department.Name
	resp.Devices = infoDevices
}

func GetDepartmentUsers(c *gin.Context, department int, areaID uint64, managerID *int) (infoUsers []departmentUser, err error) {
	var (
		users []entity.User
		owner entity.User
	)

	users, err = entity.GetDepartmentUsers(department)
	if err != nil {
		return
	}

	owner, err = entity.GetAreaOwner(areaID)
	if err != nil {
		return
	}

	for _, u := range users {
		imgUrl, _ := file.GetFileUrl(c, u.AvatarID)
		infoUser := entity.UserInfo{
			UserId:        u.ID,
			Nickname:      u.Nickname,
			IsSetPassword: u.Password != "",
			AvatarUrl:     imgUrl,
		}
		if owner.ID == u.ID {
			infoUser.RoleInfos = []entity.RoleInfo{{ID: entity.OwnerRoleID, Name: entity.Owner}}
		} else {
			infoUser.RoleInfos, err = entity.GetRoleInfos(u.ID)
		}

		isManager := managerID != nil && *managerID == u.ID
		infoUsers = append(infoUsers, departmentUser{
			UserInfo:  infoUser,
			IsManager: isManager,
		})
	}

	sort.SliceStable(infoUsers, func(i, j int) bool {
		if infoUsers[i].IsManager {
			return true
		}
		return infoUsers[i].UserId < infoUsers[j].UserId
	})
	return
}

// GetDeviceByDepartmentID 获取部门下的设备
func GetDeviceByDepartmentID(departmentID int, c *gin.Context) (infoDevices []location.Device, err error) {
	var (
		devices []entity.Device
	)
	devices, err = entity.GetDevicesByDepartmentID(departmentID)
	if err != nil {
		return
	}
	infoDevices, err = location.WrapDevices(c, devices)
	if err != nil {
		return
	}

	return
}
