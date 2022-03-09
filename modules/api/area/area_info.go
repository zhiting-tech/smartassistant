package area

import (
	errors2 "errors"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"gorm.io/gorm"
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// infoResp 家庭详情接口返回数据
type infoResp struct {
	Name          string `json:"name"`           // 区域名称
	LocationCount int64  `json:"location_count"` // 该区域的房间数量
	DepartmentCount int64  `json:"department_count"` // 该区域的房间数量
	RoleCount     int    `json:"role_count"`     // 该区域的角色数量
}

// InfoArea 用于处理区域详情接口的请求
func InfoArea(c *gin.Context) {
	var (
		err           error
		resp          infoResp
		areaId        uint64
		locationCount int64
		departmentCount int64
		area          entity.Area
		roles         []entity.Role
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	areaId, err = strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if area, err = entity.GetAreaByID(areaId); err != nil {
		return
	}
	resp.Name = area.Name

	if entity.IsHome(area.AreaType) {
		locationCount, err = entity.GetLocationCount(session.Get(c).AreaID)
	}else {
		departmentCount, err = entity.GetDepartmentCount(session.Get(c).AreaID)
	}
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			locationCount = 0
			departmentCount = 0
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}

	if roles, err = entity.GetRoles(areaId); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.RoleCount = len(roles) + 1
	resp.LocationCount = locationCount
	resp.DepartmentCount = departmentCount
	return

}
