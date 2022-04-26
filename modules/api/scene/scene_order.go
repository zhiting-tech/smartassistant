package scene

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

type OrderSceneReq struct {
	SceneIds []int `json:"scene_ids"`
}

func OrderScene(c *gin.Context) {
	var (
		req OrderSceneReq
		err error
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	u := session.Get(c)

	// 场景的id不存在,回滚数据
	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {
		count := 1
		for _, sceneId := range req.SceneIds {
			if err = entity.UpdateSceneSort(tx, sceneId, count, u.AreaID); err != nil {
				return errors.Wrap(err, errors.InternalServerErr)
			}
			count++
		}
		return nil
	}); err != nil {
		return
	}

	return
}
