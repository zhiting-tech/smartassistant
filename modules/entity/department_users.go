package entity

import (
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

type DepartmentUser struct {
	ID        int       `json:"id" gorm:"primary_key"`
	DepartmentID int    `gorm:"uniqueIndex:department_user_id"`
	Department Department `gorm:"constraint:OnDelete:CASCADE;"`
	UserID     int    `gorm:"uniqueIndex:department_user_id"`
	User 	User	`gorm:"constraint:OnDelete:CASCADE;"`
}


func (du DepartmentUser) TableName() string {
	return "department_users"
}

func CreateDepartmentUser(departmentUsers []DepartmentUser) (err error) {
	if len(departmentUsers) == 0 {
		return
	}
	if err = GetDB().Create(&departmentUsers).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

// GetUserDepartments 根据userID获取用户所在部门
func GetUserDepartments(userID int) (departments []Department, err error){
	if err = GetDB().Model(&Department{}).
		Joins("inner join department_users on departments.id=department_users.department_id").
		Where("department_users.user_id = ?", userID).Find(&departments).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

// GetDepartmentUsers 根据departmentID获取部门下用户
func GetDepartmentUsers(departmentID int) (users []User, err error) {
	if err = GetDB().Model(&User{}).
		Joins("inner join department_users on users.id=department_users.user_id").
		Where("department_users.department_id = ?", departmentID).Find(&users).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

func DelDepartmentUserByDepID(departmentID int, db *gorm.DB) (err error) {
	err = db.Where("department_id=?", departmentID).Delete(&DepartmentUser{}).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
	return
}

func DelDepartmentUserByUId(userID int, db *gorm.DB) (err error) {
	err = db.Where("user_id=?", userID).Delete(&DepartmentUser{}).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
	return
}

//UnScopedDelUserDepartments 通过用户ID删除用户与部门的关联
func UnScopedDelUserDepartments(userID int) (err error) {
	err = db.Unscoped().Where("user_id=?", userID).Delete(&DepartmentUser{}).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
	return
}

// UnScopedDelDepartmentUsers 通过部门ID删除部门与用户的关联
func UnScopedDelDepartmentUsers(departmentID int) (err error) {
	err = db.Unscoped().Where("department_id =?", departmentID).Delete(&DepartmentUser{}).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
	return
}

// WrapDepUsersOfDepID 包装部门对应用户的实体关系
func WrapDepUsersOfDepID(departmentId int, userIds []int) (departmentUsers []DepartmentUser) {
	for _, uId := range userIds {
		departmentUsers = append(departmentUsers, DepartmentUser{
			DepartmentID: departmentId,
			UserID: uId,
		})
	}
	return
}

// WrapDepUsersOfUId 包装部门对应用户的实体关系
func WrapDepUsersOfUId(userID int, departmentIds []int) (departmentUsers []DepartmentUser) {
	for _, depId := range departmentIds {
		departmentUsers = append(departmentUsers, DepartmentUser{
			DepartmentID: depId,
			UserID: userID,
		})
	}
	return
}

// GetDepartmentsByUser 获取用户所在部门信息
func GetDepartmentsByUser(user User) (departmentInfos []DepartmentInfo, err error) {
	departments, err := GetUserDepartments(user.ID)
	if err != nil {
		return
	}
	for _, department := range departments {
		isManager := false
		if isManager, err = IsDepartmentManager(user.AreaID, department.ID, user.ID); err != nil {
			return
		}
		departmentInfos = append(departmentInfos, DepartmentInfo{
			ID: department.ID,
			Name: department.Name,
			IsManager: isManager,
		})
	}
	return
}