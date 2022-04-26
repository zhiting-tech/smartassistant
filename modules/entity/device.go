package entity

import (
	"encoding/json"
	errors2 "errors"
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/types/status"

	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// Device 识别的设备
type Device struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	UniqueIdentifier string         `json:"unique_identifier" gorm:"uniqueIndex:area_id_unique_identifier"`
	PluginID         string         `json:"plugin_id" gorm:"uniqueIndex:area_id_iid_plugin_id"`
	IID              string         `json:"iid" gorm:"column:iid;uniqueIndex:area_id_iid_plugin_id;"`
	Model            string         `json:"model"`        // 型号
	Manufacturer     string         `json:"manufacturer"` // 制造商
	Type             string         `json:"type"`         // 设备类型，如：light,switch...
	CreatedAt        time.Time      `json:"created_at"`
	LocationID       int            `json:"location_id"`
	DepartmentID     int            `json:"department_id"`
	Deleted          gorm.DeletedAt `json:"deleted"`
	LogoType         *int           `json:"logo_type"`

	AreaID uint64 `json:"area_id" gorm:"type:bigint;uniqueIndex:area_id_unique_identifier;uniqueIndex:area_id_iid_plugin_id"`
	Area   Area   `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	Shadow     datatypes.JSON `json:"-"`
	ThingModel datatypes.JSON `json:"-"`
}

func (d Device) TableName() string {
	return "devices"
}

func (d *Device) AfterDelete(tx *gorm.DB) (err error) {
	// 删除设备所有相关权限
	target := types.DeviceTarget(d.ID)
	return tx.Delete(&RolePermission{}, "target = ?", target).Error
}

func (d *Device) BeforeCreate(tx *gorm.DB) (err error) {

	d.UniqueIdentifier = fmt.Sprintf("%s_%s", d.PluginID, d.IID)
	return
}

func GetDeviceByID(id int) (device Device, err error) {
	err = GetDB().First(&device, "id = ?", id).Error
	return
}

func GetDevicesByPluginID(pluginID string) (devices []Device, err error) {
	err = GetDB().Where(Device{PluginID: pluginID}).Find(&devices).Error
	return
}

// GetDeviceByIDWithUnscoped 获取设备，包括已删除
func GetDeviceByIDWithUnscoped(id int) (device Device, err error) {
	err = GetDB().Unscoped().First(&device, "id = ?", id).Error
	return
}

// GetPluginDevice 获取插件的设备
func GetPluginDevice(areaID uint64, pluginID, iid string) (device Device, err error) {
	filter := make(map[string]interface{})
	filter["iid"] = iid
	filter["plugin_id"] = pluginID

	err = GetDBWithAreaScope(areaID).Where(filter).First(&device).Error
	return
}

// GetManufacturerDevice 获取厂商的设备
func GetManufacturerDevice(areaID uint64, manufacturer, iid string) (device Device, err error) {
	filter := Device{
		IID:          iid,
		Manufacturer: manufacturer,
	}
	err = GetDBWithAreaScope(areaID).Where(filter).First(&device).Error
	return
}

func GetDevices(areaID uint64) (devices []Device, err error) {
	err = GetDBWithAreaScope(areaID).Find(&devices).Error
	return
}

// GetZhitingDevices 获取所有智汀设备
func GetZhitingDevices() (devices []Device, err error) {
	err = GetDB().Where(Device{Manufacturer: "zhiting"}).Find(&devices).Error
	return
}

func GetDevicesByLocationID(locationId int) (devices []Device, err error) {
	err = GetDB().Order("created_at asc").Find(&devices, "location_id = ?", locationId).Error
	return
}

func GetDevicesByDepartmentID(departmentId int) (devices []Device, err error) {
	err = GetDB().Order("created_at asc").Find(&devices, "department_id = ?", departmentId).Error
	return
}

func DelDeviceByIID(areaID uint64, pluginID string, iid string) (err error) {
	cond := Device{AreaID: areaID, PluginID: pluginID, IID: iid}
	err = GetDB().Delete(&Device{}, cond).Error
	return
}

func DelDeviceByID(id int) (err error) {
	d := Device{ID: id}
	err = GetDB().Delete(&d).Error
	return
}

func DelDevicesByPlgID(plgID string) (err error) {
	err = GetDB().Delete(&Device{}, "plugin_id = ?", plgID).Error
	return
}

func UpdateDevice(id int, updateDevice Device) (err error) {
	device := &Device{ID: id}
	err = GetDB().First(device).Updates(updateDevice).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.DeviceNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func GetSaDevice() (device Device, err error) {
	err = GetDB().First(&device, "model = ?", types.SaModel).Error
	return
}

func UnBindLocationDevices(locationID int) (err error) {
	err = GetDB().Find(&Device{}, "location_id = ?", locationID).Update("location_id", 0).Error
	return
}

// UnBindDepartmentDevices 解绑部门下的设备
func UnBindDepartmentDevices(departmentID int, tx *gorm.DB) (err error) {
	err = tx.Model(&Device{}).Where("department_id = ?", departmentID).Update("department_id", 0).Error
	return
}

// UnBindDepartmentDevice 解绑该设备与部门的绑定
func UnBindDepartmentDevice(deviceID int) (err error) {
	device := &Device{ID: deviceID}
	err = GetDB().First(device).Update("department_id", 0).Error
	return
}

func UnBindLocationDevice(deviceID int) (err error) {
	device := &Device{ID: deviceID}
	err = GetDB().First(device).Update("location_id", 0).Error
	return
}

// CheckSAExist SA是否已存在
func CheckSAExist(device Device, tx *gorm.DB) (err error) {
	if device.Model == types.SaModel {
		// sa设备已被绑定，直接返回
		if err = tx.First(&Device{}, "model = ? and area_id=?", types.SaModel, device.AreaID).Error; err == nil {
			return errors.Wrap(err, status.SaDeviceAlreadyBind)
		}

	}
	return nil
}

func AddDevice(d *Device, tx *gorm.DB) (err error) {
	if err = CheckSAExist(*d, tx); err != nil {
		return
	}

	if err = tx.Unscoped().Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "iid"},
			{Name: "plugin_id"},
			{Name: "area_id"},
		},
		UpdateAll: true,
	}).Create(d).Error; err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	filter := Device{
		AreaID:   d.AreaID,
		PluginID: d.PluginID,
		IID:      d.IID,
	}
	d.ID = 0
	if err = tx.First(d, filter).Error; err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}

	return
}

func UpdateThingModel(d *Device) (err error) {

	if err = GetDB().Unscoped().Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "iid"},
			{Name: "plugin_id"},
			{Name: "area_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"thing_model",
			"shadow",
			"deleted",
		}),
	}).Create(d).Error; err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	filter := Device{
		AreaID:   d.AreaID,
		PluginID: d.PluginID,
		IID:      d.IID,
	}
	d.ID = 0
	if err = GetDB().First(d, filter).Error; err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}

	return
}

// AddSADevice 添加SA设备
func AddSADevice(device *Device, tx *gorm.DB) (err error) {
	if device.Model != types.SaModel {
		return errors2.New("invalid sa")
	}

	// 初始化角色
	err = InitRole(tx, device.AreaID)
	if err != nil {
		return err
	}

	// 创建SaCreator用户和初始化权限
	var user User
	user.AreaID = device.AreaID
	// 使用同一个db，避免发生锁数据库的问题
	if err = CreateUser(&user, tx); err != nil {
		return err
	}
	if err = SetAreaOwnerID(device.AreaID, user.ID, tx); err != nil {
		return err
	}

	return AddDevice(device, tx)
}

// UpdateDeviceById 根据主键修改设备的值
func UpdateDeviceById(id int, values interface{}, tx *gorm.DB) error {
	return tx.Model(&Device{}).Where("id = ?", id).Updates(values).Error
}

// UserControlAttributes 获取用户有控制权限的属性（包括角色权限配置）
func (d Device) UserControlAttributes(up UserPermissions) (attributes []Attribute, err error) {
	tm, err := d.GetThingModel()
	if err != nil {
		return
	}

	isGatewayTm := tm.IsGateway()
	for _, instance := range tm.Instances {
		isGateway := instance.IsGateway()
		if isGatewayTm && !isGateway {
			continue
		}
		for _, srv := range instance.Services {
			// 忽略info属性
			if srv.Type == "info" {
				continue
			}
			for _, attr := range srv.Attributes {
				if !up.IsDeviceAttrControlPermit(d.ID, attr.AID) {
					continue
				}
				if attr.NoPermission() || attr.PermissionHidden() {
					continue
				}
				attributes = append(attributes, Attribute{attr})
			}
		}
	}
	return
}

// ControlAttributes 获取设备的属性（有写的权限）
func (d Device) ControlAttributes(withHidden bool) (attributes []Attribute, err error) {
	das, err := d.GetThingModel()
	if err != nil {
		return
	}

	isGatewayTm := das.IsGateway()

	for _, instance := range das.Instances {
		isGateway := instance.IsGateway()
		if isGatewayTm && !isGateway {
			continue
		}

		for _, srv := range instance.Services {
			// 忽略info属性
			if srv.Type == "info" {
				continue
			}
			for _, attr := range srv.Attributes {
				if !attr.PermissionWrite() {
					continue
				}
				if !withHidden && attr.PermissionHidden() {
					continue
				}
				attributes = append(attributes, Attribute{attr})
			}
		}
	}
	return
}

// GetShadow 从设备影子中获取属性
func (d Device) GetShadow() (shadow Shadow, err error) {
	shadow = NewShadow()
	if err = json.Unmarshal(d.Shadow, &shadow); err != nil {
		return
	}
	return
}

// GetThingModel 获取物模型，仅物模型
func (d Device) GetThingModel() (tm thingmodel.ThingModel, err error) {
	if err = json.Unmarshal(d.ThingModel, &tm); err != nil {
		return
	}
	return
}

// GetThingModelWithState 获取并包装物模型：更新值，更新权限
func (d Device) GetThingModelWithState(up UserPermissions) (tm thingmodel.ThingModel, err error) {

	tm, err = d.GetThingModel()
	if err != nil {
		return
	}
	shadow, err := d.GetShadow()
	if err != nil {
		return
	}

	// wrap attribute's value and permission
	for i := range tm.Instances {
		instance := tm.Instances[i]
		for s, srv := range instance.Services {
			for a, attr := range srv.Attributes {
				// 使用 entity.Device{}.Shadow 中的缓存值
				var val interface{}
				val, err = shadow.Get(instance.IID, attr.AID)
				if err != nil {
					return
				}
				tm.Instances[i].Services[s].Attributes[a].Val = val
				// 没有控制权限则覆盖设备属性权限
				if !up.IsDeviceAttrControlPermit(d.ID, attr.AID) {
					tm.Instances[i].Services[s].Attributes[a].SetPermissions()
				}
			}
		}
	}
	return
}
