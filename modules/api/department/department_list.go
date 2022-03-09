package department

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)


// Department 部门信息
type Department struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Sort int    `json:"sort"`
	UserCount int `json:"user_count"`
}

// departmentListResp 部门列表接口返回数据
type departmentListResp struct {
	Departments []Department `json:"departments"`
}

func ListDepartment(c *gin.Context) {
	var (
		err       error
		resp      departmentListResp
		departments []entity.Department
	)
	defer func() {
		if resp.Departments == nil {
			resp.Departments = make([]Department, 0)
		}
		response.HandleResponse(c, err, resp)
	}()

	if departments, err = entity.GetDepartments(session.Get(c).AreaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	for _, d := range departments {
		users, err := entity.GetDepartmentUsers(d.ID)
		if err != nil {
			return
		}
		resp.Departments = append(resp.Departments, Department{
			ID:   d.ID,
			Name: d.Name,
			Sort: d.Sort,
			UserCount: len(users),
		})
	}
	return
}