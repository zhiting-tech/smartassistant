package entity

import (
	errors2 "errors"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/types/status"

	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// Location 房间
type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name" gorm:"uniqueIndex:area_id_name" `
	Sort int    `json:"sort" `

	CreatedAt time.Time `json:"created_at"`

	AreaID uint64 `gorm:"type:bigint;uniqueIndex:area_id_name"`
	Area   Area   `gorm:"constraint:OnDelete:CASCADE;"`

	Deleted gorm.DeletedAt
}

func (d Location) TableName() string {
	return "locations"
}

// IsBelongsToUserArea 是否属于用户的家庭
func (d Location) IsBelongsToUserArea(user User) bool {
	return user.BelongsToArea(d.AreaID)
}

func (d *Location) BeforeCreate(tx *gorm.DB) (err error) {
	// 房间名是否重复
	if LocationNameExist(d.AreaID, d.Name) {
		err = errors.Wrap(err, status.LocationNameExist)
		return
	}

	// 添加房间, 新房间sort由后端生成
	var count int64
	if count, err = GetLocationCount(d.AreaID); err != nil {
		return
	} else {
		d.Sort = int(count) + 1
	}
	return
}

func CreateLocation(location *Location) error {
	return GetDB().Create(location).Error
}

func GetLocationByID(id int) (location Location, err error) {
	if id == 0 {
		err = errors.Wrap(err, status.LocationNotExit)
		return
	}
	err = GetDB().First(&location, "id = ?", id).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.LocationNotExit)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func LocationNameExist(areaID uint64, name string) bool {
	err := GetDBWithAreaScope(areaID).First(&Location{}, "name = ?", name).Error
	return err == nil
}

func GetLocations(areaID uint64) (locations []Location, err error) {
	err = GetDBWithAreaScope(areaID).Order("sort asc").Find(&locations).Error
	return
}

func GetLocationCount(areaID uint64) (count int64, err error) {
	err = GetDBWithAreaScope(areaID).Model(&Location{}).Count(&count).Error
	return
}

func DelLocation(id int) (err error) {
	location := &Location{ID: id}
	err = GetDB().Unscoped().First(location).Delete(location).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.LocationNotExit)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func UpdateLocation(id int, updateLocation Location) (err error) {
	location := &Location{ID: id}
	err = GetDB().First(location).Updates(updateLocation).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.LocationNotExit)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func EditLocationSort(id int, sort int) (err error) {
	err = GetDB().First(&Location{}, "id = ?", id).Update("sort", sort).Error
	return
}

func IsLocationExist(areaID uint64, locationID int) bool {
	err := GetDB().First(&Location{}, "id = ? and area_id= ?", locationID, areaID).Error
	return err == nil
}
