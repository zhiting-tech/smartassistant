package user

import (
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/extension"
	pb "github.com/zhiting-tech/smartassistant/pkg/extension/proto"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// DelUser 用于处理删除成员接口的请求
func DelUser(c *gin.Context) {
	var (
		err         error
		userID      int
		sessionUser *session.User
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	userID, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	if _, err = entity.GetUserByID(userID); err != nil {
		return
	}

	// 成员本人不能删除自己
	if sessionUser.UserID == userID {
		err = errors.Wrap(err, status.DelSelfErr)
		return
	}

	if entity.IsOwner(userID) {
		err = errors.New(status.Deny)
		return
	}

	// 删除smb数据成功，再删除用户
	if err = entity.DelUser(userID); err != nil {
		return
	}

	cloud.RemoveSAUserWithContext(c.Request.Context(), sessionUser.AreaID, userID)
	extension.GetExtensionServer().Notify(pb.SAEvent_del_user_ev, map[string]interface{}{
		"ids": []int{userID},
	})
	return

}
