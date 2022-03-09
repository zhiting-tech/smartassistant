package user

import (
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/api/area"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// userInfoResp 用户详情接口返回数据
type userInfoResp struct {
	entity.UserInfo
	IsOwner bool      `json:"is_owner"`
	IsSelf  bool      `json:"is_self"`
	Area    area.Area `json:"area"`
	DepartmentInfos []entity.DepartmentInfo `json:"department_infos,omitempty"`   // 所在部门
}

// InfoUser 用于处理用户详情接口的请求
func InfoUser(c *gin.Context) {
	var (
		err         error
		resp        userInfoResp
		user        entity.User
		userID      int
		sessionUser *session.User
		curArea 		entity.Area
	)

	defer func() {
		if err != nil {
			resp = userInfoResp{}
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

	if curArea, err = entity.GetAreaByID(user.AreaID); err != nil {
		return
	}

	resp.IsOwner = entity.IsOwner(userID)

	resp.IsSelf = userID == sessionUser.UserID
	resp.UserInfo, err = WrapUserInfo(user, resp.IsOwner)
	resp.AccountName = user.AccountName
	resp.Area, err = GetArea(curArea)

	if curArea.AreaType == entity.AreaOfCompany {
		resp.DepartmentInfos, err = entity.GetDepartmentsByUser(user)
	}

	return
}

func WrapUserInfo(user entity.User, isOwner bool) (infoUser entity.UserInfo, err error) {
	infoUser.UserId = user.ID
	infoUser.Nickname = user.Nickname
	infoUser.IsSetPassword = user.Password != ""

	if isOwner {
		infoUser.RoleInfos = []entity.RoleInfo{{ID: entity.OwnerRoleID, Name: entity.Owner}}
	} else {
		infoUser.RoleInfos, err = entity.GetRoleInfos(user.ID)
	}
	return
}

// GetArea 获取家庭信息
func GetArea(info entity.Area) (areaInfo area.Area, err error) {
	areaInfo = area.Area{
		Name: info.Name,
		ID:   strconv.FormatUint(info.ID, 10),
		AreaType: info.AreaType,
		IsBindCloud: info.IsBindCloud,
	}
	return
}
