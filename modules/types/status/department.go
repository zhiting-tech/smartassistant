package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与部门相关的响应状态码
const (
	DepartmentNotExit = iota + 10000
	DepartmentNameInputNilErr
	DepartmentNameLengthLimit
	DepartmentNameExist
)

func init() {
	errors.NewCode(DepartmentNotExit, "该部门不存在")
	errors.NewCode(DepartmentNameInputNilErr, "请输入部门名称")
	errors.NewCode(DepartmentNameLengthLimit, "部门名称长度不能超过20")
	errors.NewCode(DepartmentNameExist, "部门名称重复")
}

