package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与家庭相关的响应状态码
const (
	AreaNotExist = iota + 1000
	OwnerQuitErr
	AreaNameInputNilErr
	AreaNameLengthLimit
	SABindError
)

func init() {
	errors.NewCode(AreaNotExist, "该家庭/公司不存在")
	errors.NewCode(OwnerQuitErr, "当前%s创建者不允许退出%s")
	errors.NewCode(AreaNameInputNilErr, "请输入%s名称")
	errors.NewCode(AreaNameLengthLimit, "%s名称长度不能超过%s")
	errors.NewCode(SABindError, "SA绑定失败")
}
