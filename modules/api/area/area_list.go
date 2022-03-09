package area

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/entity"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// areaListResp 家庭列表接口返回数据
type areaListResp struct {
	Areas []Area `json:"areas"` // 家庭列表
}

// Area 家庭信息
type Area struct {
	Name string `json:"name,omitempty"` // 家庭名称
	ID   string `json:"id"`             // 家庭ID
	AreaType entity.AreaType	`json:"area_type,omitempty"`			// 该区域类型
	IsBindCloud bool `json:"is_bind_cloud,omitempty"`  // 是否绑定云端sa
}

// ListArea 用于处理家庭列表接口的请求
func ListArea(c *gin.Context) {
	var (
		err  error
		resp areaListResp
	)
	defer func() {
		if resp.Areas == nil {
			resp.Areas = make([]Area, 0)
		}
		response.HandleResponse(c, err, resp)
	}()

	area, err := entity.GetAreaByID(session.Get(c).AreaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.Areas = WrapAreas([]entity.Area{area})
	return

}

func WrapAreas(areas []entity.Area) (result []Area) {
	for _, s := range areas {
		area := Area{
			ID:   strconv.FormatUint(s.ID, 10),
			Name: s.Name,
		}
		result = append(result, area)

	}

	return result
}
