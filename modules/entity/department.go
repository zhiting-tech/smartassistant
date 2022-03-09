package entity

import (
	errors2 "errors"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type Department struct {
	ID   int    `json:"id"`
	Name string `json:"name" gorm:"uniqueIndex:dep_area_id_name" `
	Sort int    `json:"sort" `

	CreatedAt time.Time `json:"created_at"`

	AreaID uint64 `gorm:"type:bigint;uniqueIndex:dep_area_id_name"`
	Area   Area   `gorm:"constraint:OnDelete:CASCADE;"`

	ParentDepartmentID  *int
	ChildDepartments []Department  `gorm:"foreignkey:parent_department_id"`

	ManagerID  *int  `gorm:"index"`   // 主管
	Deleted   gorm.DeletedAt
}

func (d Department) TableName() string {
	return "departments"
}

type DepartmentInfo struct {
	ID   int    `json:"id,omitempty" uri:"id"`
	Name string `json:"name,omitempty"`
	IsManager bool `json:"is_manager,omitempty"`
	PId  *int 	`json:"pid,omitempty"`
}

func (d *Department) BeforeCreate(tx *gorm.DB) (err error) {
	if DepartmentNameExist(d.AreaID, d.Name) {
		err = errors.Wrap(err, status.DepartmentNameExist)
		return
	}
	var count int64
	if count, err = GetDepartmentCount(d.AreaID); err != nil {
		return
	} else {
		d.Sort = int(count) + 1
	}
	return
}

func (d Department) BeforeDelete(tx *gorm.DB) (err error) {
	if err = DelDepartmentUserByDepID(d.ID, tx); err != nil {
		return
	}
	if err = UnBindDepartmentDevices(d.ID, tx); err != nil {
		if !errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.InternalServerErr)
		} else {
			err = nil
		}
	}
	return
}

func CreateDepartment(department *Department) error {
	return GetDB().Create(department).Error
}

// GetDepartments 获取对应公司下的所有部门
func GetDepartments(areaID uint64) (departments []Department, err error) {
	err = GetDBWithAreaScope(areaID).Order("sort asc").Find(&departments).Error
	return
}

// IsDepartmentManager 该用户是否是该部门主管
func IsDepartmentManager(areaID uint64, departmentID, userID int) (isManager bool, err error) {
	filter := &Department{
		ID: departmentID,
	}
	var department Department
	err = GetDBWithAreaScope(areaID).Where(filter).Find(&department).Error
	if err != nil {
		return
	}
	if department.ManagerID == nil {
		return false, nil
	}

	return *department.ManagerID == userID, nil
}

// GetDepartmentByID 通过id获取部门
func GetDepartmentByID(id int) (department Department, err error) {
	err = GetDB().First(&department, "id = ?", id).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.DepartmentNotExit)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

// GetDepartmentCountByIds 获取通过部门id获取已有的部门数量
func GetDepartmentCountByIds(departmentIds []int) (count int64, err error) {
	if len(departmentIds) == 0 {
		return 0, nil
	}
	err = GetDB().Model(Department{}).Where("id in ?", departmentIds).Count(&count).Error
	return
}

// GetAllDepartments 获取父部门下所有子部门
func GetAllDepartments(parentDepartment []Department) (allDepartment []Department){
	allDepartment = append(allDepartment, parentDepartment...)
	for _, d := range parentDepartment {
		allDepartment = append(allDepartment, GetAllDepartments(d.ChildDepartments)...)
	}
	return allDepartment
}

// DepartmentNameExist 部门名字是否存在
func DepartmentNameExist(areaID uint64, name string) bool {
	err := GetDBWithAreaScope(areaID).First(&Department{}, "name = ?", name).Error
	return err == nil
}

// GetDepartmentCount 该公司部门数量
func GetDepartmentCount(areaID uint64) (count int64, err error) {
	err = GetDBWithAreaScope(areaID).Model(&Department{}).Count(&count).Error
	return
}

// IsDepartmentExist 是否存在该部门
func IsDepartmentExist(areaID uint64, departmentID int) bool {
	err := GetDB().First(&Department{}, "id = ? and area_id= ?", departmentID, areaID).Error
	return err == nil
}

// UpdateDepartmentSort 更新部门排序
func UpdateDepartmentSort(id int, sort int) (err error) {
	err = GetDB().First(&Department{}, "id = ?", id).Update("sort", sort).Error
	return
}

// UpdateDepartment 更新部门数据
func UpdateDepartment(id int, updateDepartment Department) (err error) {
	department := &Department{ID: id}
	err = GetDB().First(department).Updates(updateDepartment).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New(status.DepartmentNotExit)
		} else {
			err = errors.New(errors.InternalServerErr)
		}
	}
	return
}

// DelDepartment 删除部门
func DelDepartment(id int) (err error) {
	department := &Department{ID: id}
	err = GetDB().Unscoped().First(department).Delete(department).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.DepartmentNotExit)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

// GetManagerDepartments  获取主管的部门
func GetManagerDepartments(areaID uint64, managerID int) (departments []Department, err error){
	err = GetDB().Model(&Department{}).Where("manager_id = ? and area_id = ?", managerID, areaID).Find(&departments).Error
	return
}

// ResetDepartmentManager 重置部门的主管
func ResetDepartmentManager(areaID uint64, departmentID ...int) (err error){
	if len(departmentID) == 0 {
		return
	}
	return GetDB().Model(&Department{}).Where("area_id = ? and id in (?)", areaID, departmentID).Update("manager_id", 0).Error
}

// UnbindDepartmentManager 解除部门与主管关联
func UnbindDepartmentManager(managerID int, areaID uint64, tx *gorm.DB) (err error){
	return tx.Model(&Department{}).Where("manager_id = ? and area_id = ?", managerID, areaID).Update("manager_id", 0 ).Error
}