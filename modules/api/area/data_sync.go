package area

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

// DataSyncReq 数据同步接口请求参数
type DataSyncReq struct {
	Nickname string   `json:"nickname"` // 用户昵称
	Area     AreaInfo `json:"area"`     // 家庭数据
}

// AreaInfo 需要同步的家庭数据
type AreaInfo struct {
	Name      string         `json:"name"`      // 家庭名称
	Locations []LocationInfo `json:"locations"` // 家庭下的房间列表
	Departments []LocationInfo `json:"departments"` // 公司下的部门列表
}

// LocationInfo 需要同步的房间数据
type LocationInfo struct {
	Name string `json:"name"` // 房间名称
	Sort int    `json:"sort"` // 房间在房间列表中的索引
}

// DataSync 用于处理数据同步接口的请求
func DataSync(c *gin.Context) {
	var (
		err         error
		req         DataSyncReq
		sessionUser *session.User
		area        entity.Area
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	// 绑定sa后,仅允许同步一次数据
	// 根据家庭名称来判断是否已经同步过数据
	if area, err = entity.GetAreaByID(sessionUser.AreaID); err != nil {
		return
	} else {
		if area.Name != "" {
			err = errors.Wrap(err, status.AlreadyDataSync)
			return
		}
	}

	// 同步数据
	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {
		if err = tx.Model(&entity.User{}).Where("id = ?", sessionUser.UserID).Update("nickname", req.Nickname).Error; err != nil {
			return err
		}

		if err = tx.Model(&entity.Area{}).Where("id=?", sessionUser.AreaID).Update("name", req.Area.Name).Error; err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return err
		}

		for _, a := range req.Area.Locations {
			location := entity.Location{
				Name:      a.Name,
				CreatedAt: time.Now(),
				Sort:      a.Sort,
				AreaID:    sessionUser.AreaID,
			}
			if err = tx.Create(&location).Error; err != nil {
				err = errors.Wrap(err, errors.InternalServerErr)
				return err
			}
		}

		for _, d := range req.Area.Departments {
			department := entity.Department{
				Name:      d.Name,
				CreatedAt: time.Now(),
				Sort:      d.Sort,
				AreaID:    sessionUser.AreaID,
			}
			if err = tx.Create(&department).Error; err != nil {
				err = errors.Wrap(err, errors.InternalServerErr)
				return err
			}
		}

		return nil
	}); err != nil {
		err = errors.Wrap(err, status.DataSyncFail)
	}

	return
}
