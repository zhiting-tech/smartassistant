package entity

import (
	errors2 "errors"
	"time"

	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/hash"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/rand"
)

type User struct {
	// 使用自定义tag 配合 Dialector.DataTypeOf() 和 Migrator.CreateTable()确保能给字段加上autoIncrement标签
	ID          int       `json:"id" gorm:"sqliteType:integer PRIMARY KEY AUTOINCREMENT"`
	AccountName string    `json:"account_name"`
	Nickname    string    `json:"nickname"`
	Phone       string    `json:"phone"`
	Password    string    `json:"password"`
	Salt        string    `json:"salt"`
	Key         string    `json:"key" gorm:"uniqueIndex"`
	CreatedAt   time.Time `json:"created_at"`
	AvatarID    int       `json:"avatar_id"`

	AreaID             uint64 `gorm:"type:bigint;index"`
	Area               Area   `gorm:"constraint:OnDelete:CASCADE;"`
	PasswordUpdateTime time.Time
	Deleted            gorm.DeletedAt

	CommonDevices []Device `gorm:"many2many:user_common_devices"`
}

type UserCommonDevice struct {
	ID       int `gorm:"primaryKey"`
	UserID   int `gorm:"uniqueIndex:user_device"`
	DeviceID int `gorm:"uniqueIndex:user_device"`
}

type UserInfo struct {
	UserId        int        `json:"user_id"`
	RoleInfos     []RoleInfo `json:"role_infos"`
	AccountName   string     `json:"account_name"`
	Nickname      string     `json:"nickname"`
	Token         string     `json:"token,omitempty"`
	Phone         string     `json:"phone"`
	IsSetPassword bool       `json:"is_set_password"`
	AvatarUrl     string     `json:"avatar_url"`
}

func (u User) TableName() string {
	return "users"
}

func (u User) BelongsToArea(areaID uint64) bool {
	return u.AreaID == areaID
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Nickname == "" {
		u.Nickname = rand.String(rand.KindAll)
	}
	u.Key = hash.GetSaUserKey()
	u.CreatedAt = time.Now()
	if u.AreaID == 0 {
		return errors2.New("user area id is 0")
	}
	return
}

func CreateUser(user *User, tx *gorm.DB) (err error) {
	err = tx.Create(user).Error
	return
}

func GetUsers(areaID uint64) (users []User, err error) {
	err = GetDBWithAreaScope(areaID).Find(&users).Error
	return
}

func GetUserByID(id int) (user User, err error) {
	err = GetDB().Model(&User{}).First(&user, "id = ?", id).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.UserNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func GetUserByToken(token string) (user User, err error) {
	err = GetDB().Model(&User{}).First(&user, "token = ?", token).Error
	return
}

func EditUser(id int, updateUser User) (err error) {
	user := &User{ID: id}
	err = GetDB().First(user).Updates(&updateUser).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.UserNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func DelUser(id int) (err error) {
	user := &User{ID: id}
	err = GetDB().Unscoped().First(user).Delete(user).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.UserNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}
	return
}

func GetUserByAccountName(accountName string) (userInfo User, err error) {
	err = GetDB().Where("account_name = ?", accountName).First(&userInfo).Error
	return
}

func IsAccountNameExist(accountName string) bool {
	_, err := GetUserByAccountName(accountName)
	return err == nil
}

func (u User) BeforeDelete(tx *gorm.DB) (err error) {
	if err = DelUserRoleByUid(u.ID, tx); err != nil {
		return
	}
	if err = DelDepartmentUserByUId(u.ID, tx); err != nil {
		return
	}
	if err = UnbindDepartmentManager(u.ID, u.AreaID, tx); err != nil {
		return
	}
	return
}

func GetUIds(areaID uint64) (ids []int, err error) {
	if err = GetDBWithAreaScope(areaID).Model(&User{}).Pluck("id", &ids).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

func GetUserByIDAndAreaID(uID int, areaID uint64) (user User, err error) {
	err = GetDB().Model(&User{}).First(&user, "id = ? and area_id=?", uID, areaID).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.UserNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func GetUserCommonDevices(userID int) (devices []Device, err error) {

	user := User{ID: userID}
	if err = GetDB().Model(&user).Order("user_common_devices.id").Association("CommonDevices").
		Find(&devices); err != nil {
		return
	}
	return
}

// UpdateUserCommonDevices 更新用户常用设备（增加/删除/排序）
func UpdateUserCommonDevices(userID int, areaID uint64, deviceIDs []int) (err error) {

	user := User{ID: userID}
	if err = GetDB().Model(&user).Association("CommonDevices").Clear(); err != nil {
		return
	}

	if len(deviceIDs) == 0 {
		return
	}
	var devices []Device
	if err = GetDB().Where("area_id=?", areaID).Find(&devices, deviceIDs).Error; err != nil {
		return
	}
	deviceMap := make(map[int]Device)
	for _, d := range devices {
		deviceMap[d.ID] = d
	}

	devices = nil
	for _, id := range deviceIDs {
		if _, ok := deviceMap[id]; !ok {
			continue
		}
		devices = append(devices, Device{ID: id})
	}
	return GetDB().Model(&user).Association("CommonDevices").Append(devices)
}

// SetDeviceCommon 设置设备为常用设备
func SetDeviceCommon(userID int, areaID uint64, deviceID int) (err error) {

	var device Device
	if err = GetDB().Where("area_id=?", areaID).First(&device, deviceID).Error; err != nil {
		return
	}

	user := User{ID: userID}
	return GetDB().Model(&user).Association("CommonDevices").Append([]Device{device})
}

// RemoveUserCommonDevice 将设备从常用设备中取消
func RemoveUserCommonDevice(userID int, areaID uint64, deviceID int) (err error) {

	var device Device
	if err = GetDB().Where("area_id=?", areaID).First(&device, deviceID).Error; err != nil {
		return
	}

	user := User{ID: userID}
	return GetDB().Model(&user).Association("CommonDevices").Delete([]Device{device})
}

// RemoveCommonDevice 将设备从常用设备中取消
func RemoveCommonDevice(deviceID int) (err error) {
	return GetDB().Where("device_id=?", deviceID).Delete(&UserCommonDevice{}).Error
}
