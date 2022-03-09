package department

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
)

type departmentAddUserReq struct {
	Users []int `json:"users"`
}

func (req *departmentAddUserReq) validate() (err error) {
	if len(req.Users) == 0 {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	return
}

// AddDepartmentUser 处理添加/删除部门成员接口
func AddDepartmentUser(c *gin.Context) {
	var (
		err          error
		req          departmentAddUserReq
		departmentID int
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if departmentID, err = strconv.Atoi(c.Param("id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = req.validate(); err != nil {
		return
	}

	// 删除原有的部门用户关系，添加新的用户关系
	if len(req.Users) != 0 {
		if err = checkDepartmentManager(departmentID, req.Users); err != nil {
			return
		}
		if err = entity.UnScopedDelDepartmentUsers(departmentID); err != nil {
			return
		}
	}

	err = entity.CreateDepartmentUser(entity.WrapDepUsersOfDepID(departmentID, req.Users))
	return
}

// checkDepartmentManager 检查该部门主管是否被删除，并重置它
func checkDepartmentManager(departmentID int, userIDs []int) (err error) {
	department, err := entity.GetDepartmentByID(departmentID)
	if err != nil {
		return
	}
	if department.ManagerID == nil {
		return
	}
	hasManager := false
	for _, userID := range userIDs {
		if userID == *department.ManagerID {
			hasManager = true
			break
		}
	}
	if hasManager {
		return
	}
	if err = entity.ResetDepartmentManager(department.AreaID, department.ID); err != nil {
		return
	}
	return
}
