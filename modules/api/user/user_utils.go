// 提供了与用户数据相关的工具函数
package user

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

var (
	NameSizeMin = 6
	NameSizeMax = 20
)

func checkNickname(nickname string) (err error) {

	if nickname == "" || strings.TrimSpace(nickname) == "" {
		err = errors.Wrap(err, status.NickNameInputNilErr)
		return
	}

	if utf8.RuneCountInString(nickname) > NameSizeMax {
		err = errors.Wrap(err, status.NicknameLengthUpperLimit)
		return
	}
	if utf8.RuneCountInString(nickname) < NameSizeMin {
		err = errors.Wrap(err, status.NicknameLengthLowerLimit)
		return
	}

	return
}

func checkAccountName(accountName string) (err error) {

	if accountName == "" || strings.TrimSpace(accountName) == "" {
		err = errors.Wrap(err, status.AccountNameInputNilErr)
		return
	}

	if !checkAccountNameFormat(accountName) {
		err = errors.New(status.AccountNameFormatErr)
		return
	}

	if entity.IsAccountNameExist(accountName) {
		err = errors.New(status.AccountNameExist)
		return
	}

	return
}

func checkPassword(password string) (err error) {

	if password == "" || strings.TrimSpace(password) == "" {
		err = errors.Wrap(err, status.PasswordInputNilErr)
		return
	}

	if !checkPasswordFormat(password) {
		err = errors.New(status.PasswordFormatErr)
	}

	return
}

func CheckRoleID(roleID int) (err error) {
	if _, err = entity.GetRoleByID(roleID); err != nil {
		return
	}
	return
}

func checkAccountNameFormat(accountName string) bool {
	pattern := `^[\x21-\x7e]{1,}$` // 字母和数字和符号的组合 且不小于1位
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(accountName)
}

func checkPasswordFormat(password string) bool {
	pattern := `^[\x21-\x7e]{6,}$` // 字母和数字和符号的组合 且不小于6位
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(password)
}

// wrapURoles 包装用户对应的角色实体
func wrapURoles(uId int, roleIds []int) (uRoles []entity.UserRole) {
	for _, roleId := range roleIds {
		uRoles = append(uRoles, entity.UserRole{
			UserID: uId,
			RoleID: roleId,
		})
	}
	return
}
