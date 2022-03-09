package user

import (
	"github.com/zhiting-tech/smartassistant/modules/config"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	jwt2 "github.com/zhiting-tech/smartassistant/modules/utils/jwt"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// 邀请二维码过期时间
const expireAt = time.Minute * 10

// getInvitationCodeReq 获取邀请二维码接口请求参数
type getInvitationCodeReq struct {
	RoleIds []int `json:"role_ids"`
	UserId  int   `json:"-"`
	DepartmentIds []int `json:"department_ids"`
}

// getInvitationCodeResp 获取邀请二维码接口返回数据
type getInvitationCodeResp struct {
	QRCode string `json:"qr_code"`
}

// GetInvitationCode 用于处理获取邀请二维码接口的请求
func GetInvitationCode(c *gin.Context) {
	var (
		req  getInvitationCodeReq
		err  error
		resp getInvitationCodeResp
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = req.validateRequest(c); err != nil {
		return
	}
	resp, err = req.getInvitationCode(c)
}

func (req *getInvitationCodeReq) validateRequest(c *gin.Context) (err error) {
	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if len(req.RoleIds) == 0 {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	req.UserId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	u := session.Get(c)
	var curArea entity.Area
	if curArea, err = entity.GetAreaByID(u.AreaID); err != nil {
		return
	}
	if entity.IsCompany(curArea.AreaType) {
		if len(req.DepartmentIds) == 0 {
			err = errors.Wrap(err, errors.BadRequest)
			return
		}
		// 部门是否存在
		var count int64
		count, err = entity.GetDepartmentCountByIds(req.DepartmentIds)
		if err != nil {
			return
		}
		if count != int64(len(req.DepartmentIds)) {
			err = errors.Wrap(err, errors.BadRequest)
			return
		}
	}

	// 角色是否存在
	_, err = entity.GetRolesByIds(req.RoleIds)
	if err != nil {
		return
	}

	return

}

func (req getInvitationCodeReq) getInvitationCode(c *gin.Context) (resp getInvitationCodeResp, err error) {
	u := session.Get(c)
	var curArea entity.Area
	if curArea, err = entity.GetAreaByID(u.AreaID); err != nil {
		return
	}
	// 设置jwt token
	claims := jwt2.AccessClaims{
		UID:     req.UserId,
		AreaID:  u.AreaID,
		RoleIds: req.RoleIds,
		SAID:    config.GetConf().SmartAssistant.ID,
		Exp:     time.Now().Add(expireAt).Unix(),
		AreaType: curArea.AreaType,
		DepartmentIds: req.DepartmentIds,
	}

	resp.QRCode, err = jwt2.GenerateUserJwt(claims, u.Key, u.UserID)
	if err != nil {
		err = errors.Wrap(err, status.GetQRCodeErr)
	}
	return
}
