package entity

import (
	"errors"
	"unicode/utf8"

	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"

	errors2 "github.com/zhiting-tech/smartassistant/pkg/errors"

	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gorm.io/gorm"
)

const (
	OwnerRoleID = -1 // 拥有者角色的roleID
	Owner       = "拥有者"
)

type Role struct {
	ID        int
	Name      string
	IsManager bool // 管理员角色不允许修改和删除

	AreaID  uint64 `gorm:"type:bigint;index"`
	Area    Area   `gorm:"constraint:OnDelete:CASCADE;"`
	Deleted gorm.DeletedAt
}

type RoleInfo struct {
	ID   int    `json:"id,omitempty" uri:"id"`
	Name string `json:"name,omitempty"`
}

func (r Role) TableName() string {
	return "roles"

}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if utf8.RuneCountInString(r.Name) > 20 {
		err = errors2.Wrap(err, status.RoleNameLengthLimit)
	}
	return
}

func (r *Role) BeforeUpdate(tx *gorm.DB) (err error) {
	if err = tx.First(&r, r.ID).Error; err != nil {
		return
	}
	if r.IsManager {
		return errors2.New(status.Deny)
	}
	return nil
}
func (r *Role) BeforeDelete(tx *gorm.DB) (err error) {
	if err = tx.First(&r, r.ID).Error; err != nil {
		return
	}
	if r.IsManager {
		return errors2.New(status.Deny)
	}
	return nil
}
func (r *Role) AfterDelete(tx *gorm.DB) (err error) {
	return tx.Where("role_id = ?", r.ID).Delete(&UserRole{}).Error
}

func GetRoles(areaID uint64) (roles []Role, err error) {
	err = GetDBWithAreaScope(areaID).Find(&roles).Error
	return
}

func GetRolesWithTx(tx *gorm.DB, areaID uint64) (roles []Role, err error) {
	err = GetDBWithAreaScopeTx(tx, areaID).Find(&roles).Error
	return
}

func GetRoleByID(id int) (role Role, err error) {
	err = GetDB().First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors2.Wrap(err, status.RoleNotExist)
		} else {
			err = errors2.Wrap(err, errors2.InternalServerErr)
		}
	}
	return
}

func IsRoleNameExist(name string, roleID int, areaID uint64) bool {
	var db *gorm.DB
	if roleID != 0 {
		db = GetDB().Where("id != ? and name = ? and area_id=?", roleID, name, areaID)
	} else {
		db = GetDB().Where("name = ? and area_id=?", name, areaID)
	}

	err := db.First(&Role{}).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func AddRole(roleName string, areaID uint64) (Role, error) {
	return AddRoleWithDB(GetDB(), roleName, areaID)
}

func AddRoleWithDB(db *gorm.DB, roleName string, areaID uint64) (Role, error) {
	role := Role{
		Name:   roleName,
		AreaID: areaID,
	}
	err := db.FirstOrCreate(&role, role).Error
	return role, err
}

func AddManagerRoleWithDB(db *gorm.DB, roleName string, areaID uint64) (Role, error) {
	role := Role{
		Name:      roleName,
		IsManager: true,
		AreaID:    areaID,
	}
	err := db.FirstOrCreate(&role, role).Error
	return role, err
}

func UpdateRole(roleID int, roleName string) (Role, error) {
	role := Role{ID: roleID}
	err := GetDB().First(&role).Update("name", roleName).Error
	return role, err
}
func DeleteRole(roleID int) error {
	role := Role{ID: roleID}
	err := GetDB().First(&role).Delete(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors2.Wrap(err, status.RoleNotExist)
		} else {
			err = errors2.Wrap(err, errors2.InternalServerErr)
		}
	}
	return err
}

func (r *Role) AddPermissionForRole(name, action, target, attr string) error {

	p := RolePermission{
		Name:      name,
		RoleID:    r.ID,
		Action:    action,
		Target:    target,
		Attribute: attr,
	}
	return GetDB().FirstOrCreate(&p, p).Error
}
func (r *Role) AddPermissions(ps ...types.Permission) {
	r.AddPermissionsWithDB(GetDB(), ps...)
}
func (r *Role) AddPermissionsWithDB(db *gorm.DB, ps ...types.Permission) {

	for _, p := range ps {
		if err := r.addPermission(db, p); err != nil {
			logger.Error(err)
			continue
		}
	}
}

func (r *Role) addPermission(db *gorm.DB, p types.Permission) error {

	// TODO 判断是否是有效权限

	permission := RolePermission{
		Name:      p.Name,
		RoleID:    r.ID,
		Action:    p.Action,
		Target:    p.Target,
		Attribute: p.Attribute,
	}
	return db.FirstOrCreate(&permission, permission).Error
}

func (r *Role) DelPermission(p types.Permission) error {

	// TODO 判断是否是有效权限

	permission := map[string]interface{}{
		"role_id":   r.ID,
		"action":    p.Action,
		"target":    p.Target,
		"attribute": p.Attribute,
	}
	return GetDB().Where(permission).Delete(RolePermission{}).Error
}

func GetManagerRoleWithDB(db *gorm.DB) (roleInfo Role, err error) {
	err = db.Where("is_manager = ?", true).First(&roleInfo).Error
	if err != nil {
		err = errors2.Wrap(err, errors2.InternalServerErr)
		return
	}
	return
}

func InitRole(db *gorm.DB, areaId uint64) (err error) {

	var manager, member Role

	manager, err = AddManagerRoleWithDB(db, "管理员", areaId)
	if err != nil {
		return err
	}
	manager.AddPermissionsWithDB(db, types.ManagerPermission...)

	member, err = AddRoleWithDB(db, "成员", areaId)
	if err != nil {
		return err
	}
	member.AddPermissionsWithDB(db, types.MemberPermission...)

	return nil
}

func GetRolesByIds(roleIds []int) (roles []Role, err error) {
	err = GetDB().Where("id in ?", roleIds).Find(&roles).Error
	if err != nil {
		err = errors2.Wrap(err, errors2.InternalServerErr)
		return
	}
	return
}

// IsBelongsToUserArea 是否属于用户的家庭
func (r Role) IsBelongsToUserArea(user User) bool {
	return user.BelongsToArea(r.AreaID)
}

func GetRoleInfos(uID int) (roleInfos []RoleInfo, err error) {
	roles, err := GetRolesByUid(uID)
	if err != nil {
		return
	}

	for _, role := range roles {
		roleInfo := RoleInfo{
			ID:   role.ID,
			Name: role.Name,
		}
		roleInfos = append(roleInfos, roleInfo)
	}
	return
}
