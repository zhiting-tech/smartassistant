package user

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
)

// UserInfoDepartmentResp 用户所在部门
type UserInfoDepartmentResp struct {
	DepartmentInfos []entity.DepartmentInfo `json:"department_infos"`   // 所在部门
}

// UserInfoDepartment 用于处理用户详情接口的请求
func UserInfoDepartment(c *gin.Context) {
	var (
		err         error
		resp        UserInfoDepartmentResp
		user        entity.User
		userID      int
		sessionUser *session.User
	)

	defer func() {
		if resp.DepartmentInfos == nil {
			resp.DepartmentInfos = make([]entity.DepartmentInfo, 0)
		}

		response.HandleResponse(c, err, &resp)
	}()

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	userID, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if user, err = entity.GetUserByID(userID); err != nil {
		return
	}
	resp.DepartmentInfos, err = entity.GetDepartmentsByUser(user)
	return
}