package entity

import (
	"encoding/json"
	errors2 "errors"
	"fmt"
	"time"

	"github.com/mozillazg/go-unidecode"
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
	ID        int    `json:"id"`
	ParentIID string `json:"parent_iid" gorm:"column:parent_iid"`
	Name      string `json:"name"`
	Pinyin    string `json:"pinyin"`

	UniqueIdentifier string         `json:"unique_identifier" gorm:"uniqueIndex:area_id_unique_identifier"`
	PluginID         string         `json:"plugin_id" gorm:"uniqueIndex:area_id_iid_plugin_id"`
	IID              string         `json:"iid" gorm:"column:iid;uniqueIndex:area_id_iid_plugin_id;"`
	Model            string         `json:"model"`        // 型号
	Manufacturer     string         `json:"manufacturer"` // 制造商
	Type             string         `json:"type"`         // 设备类型，如：light,switch...
	CreatedAt        time.Time      `json:"created_at"`
	Deleted          gorm.DeletedAt `json:"deleted"`
	LogoType         *int           `json:"logo_type"`

	Shadow     datatypes.JSON `json:"-"`
	ThingModel datatypes.JSON `json:"-"`

	SyncData string `json:"-"` // 自定义的客户端同步信息

	AreaID    uint64 `json:"area_id" gorm:"type:bigint;uniqueIndex:area_id_unique_identifier;uniqueIndex:area_id_iid_plugin_id"`
	Area      Area   `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	AreaOrder int    `gorm:"default:0"` // 设备在家庭中的排序

	// 房间
	LocationID    int `json:"location_id"`
	LocationOrder int // 设备在房间中的排序

	// 部门
	DepartmentID    int `json:"department_id"`
	DepartmentOrder int // 设备在部门中的排序
}

func (d Device) TableName() string {
	return "devices"
}

func (d Device) IsSa() bool {
	return d.Model == types.SaModel
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

func (d *Device) BeforeUpdate(tx *gorm.DB) (err error) {

	if tx.Statement.Changed("LocationID") {
		tx.Select("LocationOrder").Statement.SetColumn("LocationOrder", 0)
	}
	if tx.Statement.Changed("DepartmentID") {
		tx.Select("DepartmentOrder").Statement.SetColumn("DepartmentOrder", 0)
	}
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

// GetPluginDevices 获取插件的所有设备
func GetPluginDevices(areaID uint64, pluginID string) (devices []Device, err error) {
	filter := make(map[string]interface{})
	filter["plugin_id"] = pluginID

	err = GetDBWithAreaScope(areaID).Where(filter).Find(&devices).Error
	return
}

func GetDevices(areaID uint64) (devices []Device, err error) {
	err = GetDBWithAreaScope(areaID).Find(&devices).Error
	return
}

func GetDevicesOrderByPinyin(areaID uint64) (devices []Device, err error) {
	err = GetDBWithAreaScope(areaID).Order("lower(pinyin)").Find(&devices).Error
	return
}

func GetOrderDevices(areaID uint64) (devices []Device, err error) {
	err = GetDBWithAreaScope(areaID).Order("area_order desc,created_at").
		Find(&devices).Error
	return
}

func GetDevicesByLocationID(locationId int) (devices []Device, err error) {
	err = GetDB().Order("created_at asc").
		Find(&devices, "location_id = ?", locationId).Error
	return
}

func GetDevicesByDepartmentID(departmentId int) (devices []Device, err error) {
	err = GetDB().Order("created_at asc").
		Find(&devices, "department_id = ?", departmentId).Error
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
	d, err := GetDeviceByID(id)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.DeviceNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}
	err = GetDB().Model(&d).Updates(updateDevice).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
	return
}

// IsDeviceExist 设备是否存在
func IsDeviceExist(areaID uint64, pluginID, iid string) (isExist bool, err error) {

	// 网关未添加则设备不更新
	if _, err = GetPluginDevice(areaID, pluginID, iid); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GetSaDevice() (device Device, err error) {
	err = GetDB().First(&device, "model = ?", types.SaModel).Error
	return
}

func UnBindLocationDevices(locationID int) (err error) {
	err = GetDB().Find(&Device{}, "location_id = ?", locationID).
		Update("location_id", 0).Error
	return
}

// UnBindDepartmentDevices 解绑部门下的设备
func UnBindDepartmentDevices(departmentID int, tx *gorm.DB) (err error) {
	err = tx.Model(&Device{}).Where("department_id = ?", departmentID).
		Update("department_id", 0).Error
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

func ReorderDevices(areaID uint64, deviceIDs []int) (err error) {

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
	length := len(deviceIDs)
	for i, id := range deviceIDs {
		if _, ok := deviceMap[id]; !ok {
			continue
		}
		devices = append(devices, Device{
			ID:        id,
			AreaOrder: length - i})
	}
	return GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"area_order"}),
	}).Create(&devices).Error
}

// ReorderLocationDevices 重新排序房间设备
func ReorderLocationDevices(locationID int, deviceIDs []int) (err error) {

	if len(deviceIDs) == 0 {
		return
	}
	var devices []Device
	if err = GetDB().Where("location_id=?", locationID).
		Find(&devices, deviceIDs).Error; err != nil {
		return
	}
	deviceMap := make(map[int]Device)
	for _, d := range devices {
		deviceMap[d.ID] = d
	}

	devices = nil
	length := len(deviceIDs)
	for i, id := range deviceIDs {
		if _, ok := deviceMap[id]; !ok {
			continue
		}
		devices = append(devices, Device{
			ID:            id,
			LocationID:    locationID,
			LocationOrder: length - i})
	}

	return GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"location_id", "location_order"}),
	}).Create(&devices).Error
}

// ReorderDepartmentDevices 重新排序部门设备
func ReorderDepartmentDevices(departmentID int, deviceIDs []int) (err error) {

	if len(deviceIDs) == 0 {
		return
	}
	var devices []Device
	if err = GetDB().Where("department_id=?", departmentID).
		Find(&devices, deviceIDs).Error; err != nil {
		return
	}
	deviceMap := make(map[int]Device)
	for _, d := range devices {
		deviceMap[d.ID] = d
	}

	devices = nil
	length := len(deviceIDs)
	for i, id := range deviceIDs {
		if _, ok := deviceMap[id]; !ok {
			continue
		}
		devices = append(devices, Device{
			ID:              id,
			DepartmentID:    departmentID,
			DepartmentOrder: length - i})
	}

	return GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"department_id", "department_order"}),
	}).Create(&devices).Error
}

func GetOrderDepartmentDevices(departmentID int) (devices []Device, err error) {

	var department Department
	if department, err = GetDepartmentByID(departmentID); err != nil {
		return
	}
	if err = GetDB().Where("department_id=?", department.ID).
		Order("department_order desc,created_at").Find(&devices).Error; err != nil {
		return
	}
	return
}

func GetOrderLocationDevices(locationID int) (devices []Device, err error) {

	var location Location
	if location, err = GetLocationByID(locationID); err != nil {
		return
	}

	if err = GetDB().Where("location_id=?", location.ID).
		Order("location_order desc,created_at").Find(&devices).Error; err != nil {
		return
	}
	return
}

// CheckSAExist SA是否已存在
func CheckSAExist(device Device, tx *gorm.DB) (err error) {
	if device.IsSa() {
		// sa设备已被绑定，直接返回
		if err = tx.First(&Device{}, "model = ? and area_id=?", types.SaModel, device.AreaID).Error; err == nil {
			return errors.Wrap(err, status.SaDeviceAlreadyBind)
		}

	}
	return nil
}

func CreateDevice(d *Device, tx *gorm.DB) (err error) {

	d.Pinyin = unidecode.Unidecode(d.Name)
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
	// CreatedAt 在上面的upsert代码中没法更新成功
	if err = tx.Model(d).Where("id=?", d.ID).Update("CreatedAt", time.Now()).Error; err != nil {
		return
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
			"parent_iid",
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

// CreateSA 添加SA设备
func CreateSA(device *Device, tx *gorm.DB) (err error) {
	if !device.IsSa() {
		return errors2.New("invalid sa")
	}
	if err = CheckSAExist(*device, tx); err != nil {
		return
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

	return CreateDevice(device, tx)
}

// UserAttributes 获取设备所有有权限的属性
func (d Device) UserAttributes() (attributes []Attribute, err error) {

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
				if attr.NoPermission() || attr.PermissionHidden() {
					continue
				}
				attributes = append(attributes, Attribute{srv.Type, attr})
			}
		}
	}
	return
}

// TriggerableAttributes 获取设备所有可以作为触发条件的属性
func (d Device) TriggerableAttributes() (attributes []Attribute, err error) {

	attrs, err := d.UserAttributes()
	if err != nil {
		return
	}
	for _, attr := range attrs {
		if !attr.PermissionSceneHidden() && (attr.PermissionRead() || attr.PermissionNotify()) {
			attributes = append(attributes, attr)
		}
	}
	return
}

// ControllableAttributes 获取设备所有可以控制的属性
func (d Device) ControllableAttributes(up UserPermissions) (attributes []Attribute, err error) {

	attrs, err := d.UserAttributes()
	if err != nil {
		return
	}
	for _, attr := range attrs {
		if attr.PermissionWrite() && up.IsDeviceAttrControlPermit(d.ID, attr.AID) && !attr.PermissionSceneHidden() {
			attributes = append(attributes, attr)
		}
	}
	return
}

// IsTriggerable 判断设备可以作为触发条件被选择
func (d Device) IsTriggerable() bool {
	attributes, err := d.TriggerableAttributes()
	if err != nil {
		return false
	}

	return len(attributes) != 0
}

// IsControllable 判断设备可以作为执行任务被选择
func (d Device) IsControllable(up UserPermissions) bool {
	attributes, err := d.ControllableAttributes(up)
	if err != nil {
		return false
	}
	return len(attributes) != 0
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
				attributes = append(attributes, Attribute{srv.Type, attr})
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
	if d.PluginID == "" { // 仅为客户端同步的设备没有物模型
		return
	}
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
				if !up.IsDeviceAttrControlPermit(d.ID, attr.AID) && !attr.PermissionHidden() {
					tm.Instances[i].Services[s].Attributes[a].
						RemovePermissions(thingmodel.AttributePermissionWrite)
				}
			}
		}
	}
	return
}
