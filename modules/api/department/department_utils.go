package department

import (
	"strings"
	"unicode/utf8"

	"github.com/zhiting-tech/smartassistant/modules/types/status"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

func checkDepartmentName(name string) (err error) {

	if name == "" || strings.TrimSpace(name) == "" {
		err = errors.Wrap(err, status.DepartmentNameInputNilErr)
		return
	}

	if utf8.RuneCountInString(name) > 20 {
		err = errors.Wrap(err, status.DepartmentNameLengthLimit)
		return
	}

	return
}
