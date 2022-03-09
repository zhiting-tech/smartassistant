package department

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/extension"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	pb "github.com/zhiting-tech/smartassistant/pkg/extension/proto"
	"strconv"
)

// DelDepartment 用于处理删除部门接口的请求
func DelDepartment(c *gin.Context) {
	var (
		departmentId  int
		err error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	departmentId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = entity.DelDepartment(departmentId); err != nil {
		return
	}
	extension.GetExtensionServer().Notify(pb.SAEvent_del_department_ev, map[string]interface{}{
		"ids": []int{departmentId},
	})
	return
}
