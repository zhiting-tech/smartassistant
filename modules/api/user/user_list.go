package user

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"sort"
)

// ListUserResp 成员列表接口返回数据
type ListUserResp struct {
	IsOwner   bool              `json:"is_owner"`
	UserCount int               `json:"user_count"`
	Users     []entity.UserInfo `json:"users"`
}

// ListUser 用于处理成员列表接口的请求
func ListUser(c *gin.Context) {
	var (
		err         error
		resp        ListUserResp
		sessionUser *session.User
	)

	defer func() {
		if err != nil {
			resp = ListUserResp{}
		}

		if len(resp.Users) == 0 {
			resp.Users = make([]entity.UserInfo, 0)
		}
		response.HandleResponse(c, err, &resp)
	}()

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	resp.IsOwner = entity.IsOwner(sessionUser.UserID)

	userRoles, err := entity.GetUserRoles(sessionUser.AreaID)
	if err != nil {
		return
	}

	resp.Users, err = WrapUsers(userRoles, sessionUser.AreaID)
	resp.UserCount = len(resp.Users)
	return

}

func WrapUsers(userRoles []entity.UserRole, areaID uint64) (listUsers []entity.UserInfo, err error) {

	users := make(map[int]entity.UserInfo)

	for _, userRole := range userRoles {

		if v, ok := users[userRole.UserID]; !ok {
			var userInfo entity.UserInfo
			userInfo.UserId = userRole.UserID
			userInfo.Nickname = userRole.User.Nickname
			userInfo.IsSetPassword = userRole.User.Password != ""
			userInfo.RoleInfos = []entity.RoleInfo{{ID: userRole.Role.ID, Name: userRole.Role.Name}}
			users[userRole.UserID] = userInfo
		} else {
			roleInfo := entity.RoleInfo{ID: userRole.Role.ID, Name: userRole.Role.Name}
			v.RoleInfos = append(v.RoleInfos, roleInfo)
			users[userRole.UserID] = v
		}
	}

	for _, user := range users {
		listUsers = append(listUsers, user)
	}

	owner, err := entity.GetAreaOwner(areaID)
	if err != nil {
		return
	}
	ownerInfo := entity.UserInfo{
		UserId:        owner.ID,
		RoleInfos:     []entity.RoleInfo{{ID: entity.OwnerRoleID, Name: entity.Owner}},
		AccountName:   owner.AccountName,
		Nickname:      owner.Nickname,
		IsSetPassword: owner.Password != "",
	}
	listUsers = append(listUsers, ownerInfo)
	// 返回的成员列表按照加入家庭的时间正序排序，这里使用UserID进行排序
	sort.SliceStable(listUsers, func(i, j int) bool {
		return listUsers[i].UserId < listUsers[j].UserId
	})
	return
}
