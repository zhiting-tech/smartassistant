package department

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"time"
)

// departmentAddReq 添加部门接口请求参数
type departmentAddReq struct {
	Name string `json:"name"`
}

func (req *departmentAddReq) Validate() (err error) {
	if err = checkDepartmentName(req.Name); err != nil {
		return
	}
	return
}

// AddDepartment 用于处理添加部门接口的请求
func AddDepartment(c *gin.Context) {
	var (
		newDepartment entity.Department
		req         departmentAddReq
		err         error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = req.Validate(); err != nil {
		return
	}
	newDepartment.Name = req.Name
	newDepartment.CreatedAt = time.Now()
	newDepartment.AreaID = session.Get(c).AreaID

	if err = entity.CreateDepartment(&newDepartment); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}
