package user

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

func TransferOwner(c *gin.Context) {
	var (
		err error
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	user := session.Get(c)
	if user == nil {
		err = errors.Wrap(err, status.InvalidUserCredentials)
		return
	}

	// 判断是否是拥有者
	if !user.IsOwner {
		err = errors.Wrap(err, status.Deny)
		return
	}

	newOwnerID, err := strconv.Atoi(c.Param("id"))
	if err != nil || newOwnerID == user.UserID {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	_, err = entity.GetUserByID(newOwnerID)
	if err != nil {
		return
	}

	// 转移角色
	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {
		// 删除新拥有者的用户与角色信息
		if err = tx.Unscoped().Where("user_id=?", newOwnerID).Delete(&entity.UserRole{}).Error; err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return err
		}

		// 更新area的拥有者
		if err = tx.Model(&entity.Area{}).Where("id=?", user.AreaID).Update("owner_id", newOwnerID).Error; err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return err
		}

		roleManager, err := entity.GetManagerRoleWithDB(tx)
		if err != nil {
			return err
		}

		// 添加旧拥有者为管理员
		uRole := entity.UserRole{
			UserID: user.UserID,
			RoleID: roleManager.ID,
		}

		if err := tx.Create(&uRole).Error; err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return err
		}

		return nil
	}); err != nil {
		return
	}

}
