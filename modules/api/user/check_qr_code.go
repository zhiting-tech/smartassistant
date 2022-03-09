package user

import (
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/api/area"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	jwt2 "github.com/zhiting-tech/smartassistant/modules/utils/jwt"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// checkQrCodeReq 扫描邀请二维码接口请求参数
type checkQrCodeReq struct {
	QrCode   string `json:"qr_code"`
	Nickname string `json:"nickname"`
	roleIds  []int
	areaId   uint64
	areaType entity.AreaType
	departmentIds []int
}

// CheckQrCodeResp 扫描邀请二维码接口返回数据
type CheckQrCodeResp struct {
	UserInfo entity.UserInfo `json:"user_info"`
	AreaInfo area.Area       `json:"area_info"`
}

func (req *checkQrCodeReq) validateRequest(c *gin.Context) (err error) {
	var curArea entity.Area
	if err = c.BindJSON(&req); err != nil {
		return
	}

	//	二维码是否在有效时间
	claims, err := jwt2.ValidateUserJwt(req.QrCode)
	if err != nil {
		//	二维码是否在有效时间
		if err.Error() == jwt2.ErrTokenIsExpired.Error() {
			return errors.New(status.QRCodeExpired)
		}
		err = errors.Wrap(err, status.QRCodeInvalid)
		return
	}

	// 判断是否是拥有者
	u := session.Get(c)
	if u != nil {
		if entity.IsOwnerOfArea(u.UserID, claims.AreaID) {
			err = errors.New(status.OwnerForbidJoinAreaAgain)
			return
		}
	}

	// 判断二维码创建者是否有生成二维码权限
	var creatorID = claims.UID
	if !entity.JudgePermit(creatorID, types.AreaGetCode) {
		err = errors.New(status.QRCodeCreatorDeny)
		return
	}

	req.areaId = claims.AreaID
	// 对应家庭未删除
	curArea, err = entity.GetAreaByID(req.areaId)
	if err != nil {
		return
	}
	// 角色未被删除
	req.roleIds = claims.RoleIds
	roles, err := entity.GetRolesByIds(req.roleIds)
	if err != nil {
		return
	}

	if len(roles) != len(req.roleIds) {
		err = errors.New(status.RoleNotExist)
		return
	}
	// 区域类型是否一致
	req.areaType = claims.AreaType
	if curArea.AreaType != req.areaType {
		err = errors.New(status.AreaTypeNotEqual)
		return
	}

	// 部门未被删除
	// TODO 这里实现不好，尝试用hook去做判断，但要注意是否使用事务完成扫码逻辑
	req.departmentIds = claims.DepartmentIds
	if entity.IsCompany(curArea.AreaType) {
		var departmentCount int64
		departmentCount, err = entity.GetDepartmentCountByIds(req.departmentIds)
		if err != nil {
			return
		}
		if departmentCount != int64(len(req.departmentIds)){
			err = errors.New(status.DepartmentNotExit)
			return
		}
	}

	return
}

// CheckQrCode 用于处理扫描邀请二维码接口的请求
func CheckQrCode(c *gin.Context) {
	var (
		req  checkQrCodeReq
		err  error
		resp CheckQrCodeResp
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = req.validateRequest(c); err != nil {
		return
	}

	resp, err = req.checkQrCode(c)
	if err != nil {
		return
	}

}

func (req *checkQrCodeReq) checkQrCode(c *gin.Context) (resp CheckQrCodeResp, err error) {
	u := session.GetUserByToken(c)

	var (
		uRoles []entity.UserRole
		uDepartments []entity.DepartmentUser
		currentArea entity.Area
	)

	if currentArea, err = entity.GetAreaByID(req.areaId); err != nil {
		return
	}

	var user entity.User
	if u == nil {
		// 未加入该家庭
		user = entity.User{
			Nickname: req.Nickname,
			AreaID:   req.areaId,
		}
		if err = entity.CreateUser(&user, entity.GetDB()); err != nil {
			return
		}
		uRoles = wrapURoles(user.ID, req.roleIds)
		uDepartments = entity.WrapDepUsersOfUId(user.ID, req.departmentIds)

	} else {
		user, err = entity.GetUserByID(u.UserID)
		if err != nil {
			return
		}

		// 重复扫码，以最后扫码角色为主
		// 删除用户原有角色,部门关系
		if err = entity.UnScopedDelURoleByUid(u.UserID); err != nil {
			return
		}

		if len(req.departmentIds) > 0 && entity.IsCompany(currentArea.AreaType) {
			if err = CheckDepartmentsManager(u.UserID, req.departmentIds, req.areaId); err != nil {
				return
			}
			uDepartments = entity.WrapDepUsersOfUId(user.ID, req.departmentIds)
		}
		uRoles = wrapURoles(user.ID, req.roleIds)
	}
	// 给用户创建角色
	if err = entity.CreateUserRole(uRoles); err != nil {
		return
	}

	if len(uDepartments) > 0 && entity.IsCompany(currentArea.AreaType) {
		if err = entity.CreateDepartmentUser(uDepartments); err != nil {
			return
		}
	}

	resp.UserInfo = entity.UserInfo{
		UserId:        user.ID,
		AccountName:   user.AccountName,
		Nickname:      user.Nickname,
		IsSetPassword: user.Password != "",
		Phone:         user.Phone,
	}

	if u != nil {
		resp.UserInfo.Token = u.Token
	} else {
		resp.UserInfo.Token, err = oauth.GetSAUserToken(user, c.Request)
		if err != nil {
			return
		}

	}

	resp.AreaInfo = area.Area{
		ID: strconv.FormatUint(req.areaId, 10),
	}

	return
}
