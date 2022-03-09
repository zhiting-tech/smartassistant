package department

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
)

type departmentUpdateReq struct {
	Name	*string   `json:"name"`
	ManagerID *int	 `json:"manager_id"`
}

func (req *departmentUpdateReq) validate(departmentId int) (updateDepartment entity.Department, err error) {
	var (
		department entity.Department
	)
	if department, err = entity.GetDepartmentByID(departmentId); err != nil {
		return
	}
	if req.Name != nil {
		if err = checkDepartmentName(*req.Name); err != nil {
			return
		}

		if department.Name != *req.Name {
			updateDepartment.Name = *req.Name
		}
	}

	if req.ManagerID != nil  {
		if *req.ManagerID != 0 {
			if _, err = entity.GetUserByID(*req.ManagerID); err != nil {
				return
			}
		}
		updateDepartment.ManagerID = req.ManagerID
	}
	return
}

// UpdateDepartment 处理部门更新
func UpdateDepartment(c *gin.Context) {
	var (
		err  	error
		req 	departmentUpdateReq
		departmentId int
		updateDepartment  entity.Department
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if departmentId, err = strconv.Atoi(c.Param("id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if updateDepartment, err = req.validate(departmentId); err != nil {
		return
	}

	if err = entity.UpdateDepartment(departmentId, updateDepartment); err != nil {
		return
	}
	return
}