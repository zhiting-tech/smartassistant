package area

import (
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"strings"
	"unicode/utf8"

	"github.com/zhiting-tech/smartassistant/modules/types/status"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

const (
	homeLenLimit = 30
	companyLimit = 50
)

func checkAreaName(name string, areaType entity.AreaType) (err error) {
	lenLimit := companyLimit
	areaModeStr := areaType.String()
	if entity.IsHome(areaType) {
		lenLimit = homeLenLimit

	}
	if name == "" || strings.TrimSpace(name) == "" {
		err = errors.Wrapf(err, status.AreaNameInputNilErr, areaModeStr)
		return
	}

	if utf8.RuneCountInString(name) > lenLimit {
		err = errors.Wrapf(err, status.AreaNameLengthLimit, areaModeStr, lenLimit)
		return
	}
	return
}
