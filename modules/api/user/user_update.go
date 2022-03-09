package user

import (
	"github.com/zhiting-tech/smartassistant/modules/utils/hash"
	"github.com/zhiting-tech/smartassistant/pkg/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// updateUserReq 修改用户接口请求参数
type updateUserReq struct {
	Nickname    *string `json:"nickname"`
	AccountName *string `json:"account_name"`
	Password    *string `json:"password"`
	OldPassword  *string `json:"old_password"`
	RoleIds     []int   `json:"role_ids"`
	DepartmentIds []int `json:"department_ids,omitempty"`
}

func (req *updateUserReq) Validate(updateUid, loginId int, areaType entity.AreaType) (updateUser entity.User, err error) {
	if len(req.RoleIds) != 0 {
		// 判断是否有修改角色权限
		if !entity.JudgePermit(loginId, types.AreaUpdateMemberRole) {
			err = errors.Wrap(err, status.Deny)
			return
		}
	}

	if entity.IsCompany(areaType) && req.DepartmentIds != nil {
		if !entity.JudgePermit(loginId, types.AreaUpdateMemberDepartment) {
			err = errors.Wrap(err, status.Deny)
			return
		}
	}

	// 自己才允许修改自己的用户名,密码和昵称
	if req.Nickname != nil || req.AccountName != nil || req.Password != nil || req.OldPassword != nil{
		if loginId != updateUid {
			err = errors.New(status.Deny)
			return
		}
	}

	if req.Nickname != nil {
		if err = checkNickname(*req.Nickname); err != nil {
			return
		} else {
			updateUser.Nickname = *req.Nickname
		}
	}
	if req.AccountName != nil {
		if err = checkAccountName(*req.AccountName); err != nil {
			return
		} else {
			updateUser.AccountName = *req.AccountName
		}
	}

	if req.Password != nil {
		if err = checkPassword(*req.Password); err != nil {
			return
		}
		var userInfo entity.User
		userInfo, err = entity.GetUserByID(loginId)
		// 通过密码是否设置过，判断是否是修改密码还是初始设置密码
		if userInfo.Password == "" {
			salt := rand.String(rand.KindAll)
			updateUser.Salt = salt
			hashNewPassword := hash.GenerateHashedPassword(*req.Password, salt)
			updateUser.Password = hashNewPassword
		}else {
			if userInfo.Password != hash.GenerateHashedPassword(*req.OldPassword, userInfo.Salt) {
				err = errors.New(status.OldPasswordErr)
				return
			}
			updateUser.Password = hash.GenerateHashedPassword(*req.Password, userInfo.Salt)
			updateUser.PasswordUpdateTime = time.Now()
		}
	}

	return
}

// UpdateUser 用于处理修改用户接口的请求
func UpdateUser(c *gin.Context) {
	var (
		err         error
		req         updateUserReq
		updateUser  entity.User
		sessionUser *session.User
		userID      int
		curArea  		entity.Area
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if userID, err = strconv.Atoi(c.Param("id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if curArea, err = entity.GetAreaByID(sessionUser.AreaID); err != nil {
		return
	}

	if updateUser, err = req.Validate(userID, sessionUser.UserID, curArea.AreaType); err != nil {
		return
	}

	if len(req.RoleIds) != 0 {
		if entity.IsOwner(userID) {
			err = errors.New(status.NotAllowModifyRoleOfTheOwner)
			return
		}
		// 删除用户原有角色
		if err = entity.UnScopedDelURoleByUid(userID); err != nil {
			return
		}
		if err = entity.CreateUserRole(wrapURoles(userID, req.RoleIds)); err != nil {
			return
		}
	}

	if entity.IsCompany(curArea.AreaType) && req.DepartmentIds != nil{
		if err = CheckDepartmentsManager(userID, req.DepartmentIds, curArea.ID); err != nil {
			return
		}
		if err = entity.CreateDepartmentUser(entity.WrapDepUsersOfUId(userID, req.DepartmentIds)); err != nil {
			return
		}
	}

	if err = entity.EditUser(userID, updateUser); err != nil {
		return
	}

	return
}

// getDelManagerDepartments 获取需要重置主管的部门
func getDelManagerDepartments(oldDepartment []entity.Department, newDepartmentIDs []int) (departmentIDs []int){
	if len(newDepartmentIDs) == 0 {
		for _, department:= range oldDepartment {
			departmentIDs = append(departmentIDs, department.ID)
		}
		return
	}

	for _, department := range oldDepartment {
		isOld := true
		for _, departmentID := range newDepartmentIDs {
			if departmentID == department.ID {
				isOld = false
				break
			}
		}
		if isOld {
			departmentIDs = append(departmentIDs, department.ID)
		}
	}
	return
}

// CheckDepartmentsManager 检查多个部门的主管是否被删除，并重置它
// TODO 尝试在删除的hook中做删除主管的操作，但要注意，原先的删除/更新逻辑的影响
func CheckDepartmentsManager(userID int, departmentIds []int, areaID uint64) (err error){
	var departments []entity.Department
	if departments, err = entity.GetManagerDepartments(areaID, userID); err != nil {
		return
	}
	delManagerDepartmentIDs := getDelManagerDepartments(departments, departmentIds)
	if err = entity.ResetDepartmentManager(areaID, delManagerDepartmentIDs...); err != nil {
		return
	}

	if err = entity.UnScopedDelUserDepartments(userID); err != nil {
		return
	}
	return
}
