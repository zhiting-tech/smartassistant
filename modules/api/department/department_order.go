package department

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

// departmentOrderReq 调整部门列表顺序接口请求体
type departmentOrderReq struct {
	DepartmentsOrder []int `json:"departments_id"`
}

// DepartmentOrder 用于处理调整部门列表顺序的请求
func DepartmentOrder(c *gin.Context) {
	var (
		err   error
		req   departmentOrderReq
		count int64
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	u := session.Get(c)
	if count, err = entity.GetDepartmentCount(u.AreaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	} else if len(req.DepartmentsOrder) != int(count) {
		err = errors.New(errors.BadRequest)
		return
	}

	// department_id不存在,回滚数据
	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {
		for i, departmentId := range req.DepartmentsOrder {
			if !entity.IsDepartmentExist(u.AreaID, departmentId) {
				err = errors.Wrap(err, status.LocationNotExit)
				return err
			}
			if err = entity.UpdateDepartmentSort(departmentId, i+1); err != nil {
				err = errors.Wrap(err, errors.InternalServerErr)
				return err
			}
		}
		return nil
	}); err != nil {
		return
	}

	return
}
