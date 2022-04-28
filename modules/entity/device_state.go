package entity

import (
	"time"

	"gorm.io/datatypes"
)

type DeviceState struct {
	ID        int            `json:"id"`
	DeviceID  int            `json:"device_id"`
	PluginID  string         `json:"plugin_id"`
	IID       string         `json:"iid" gorm:"column:iid"`
	State     datatypes.JSON `json:"context"` // refer to event.State
	CreatedAt time.Time      `json:"created_at"`
}

func (d DeviceState) MarshalJSON() ([]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (d DeviceState) TableName() string {
	return "device_states"
}

func InsertDeviceState(d Device, state []byte) (err error) {
	deviceState := &DeviceState{
		DeviceID: d.ID,
		PluginID: d.PluginID,
		IID:      d.IID,
		State:    state,
	}
	err = GetDB().Create(deviceState).Error
	if err != nil {
		return
	}
	return
}

func GetDeviceStates(deviceID int, attrType *string, size int, index *int, startAt, endAt *int64) (states []DeviceState, err error) {
	db := GetDB().Where("device_id=?", deviceID).Order("id desc")
	if attrType != nil {
		attrQuery := datatypes.JSONQuery("state").
			Equals(*attrType, "type")
		db.Clauses(attrQuery)
	}
	if index != nil {
		db.Where("id<?", index)
	}
	if startAt != nil {
		db.Where("created_at > ?", time.Unix(*startAt, 0))
	}
	if endAt != nil {
		db.Where("created_at < ?", time.Unix(*endAt, 0))
	}
	err = db.Limit(size).Find(&states).Error
	return
}

func GetPluginDeviceStates(pluginID, iid string) (states []DeviceState, err error) {
	err = GetDB().Where(DeviceState{PluginID: pluginID, IID: iid}).Find(&states).Error
	return
}
