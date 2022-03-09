package event

import (
	"encoding/json"
	"github.com/zhiting-tech/smartassistant/modules/entity"
)

type EventType string

const (
	AttributeChange EventType = "attribute_change"
	DeviceIncrease  EventType = "device_increase"
	DeviceDecrease  EventType = "device_decrease"
)

type EventMessage struct {
	EventType EventType
	AreaID    uint64
	Param     map[string]interface{}
}

func NewEventMessage(et EventType, areaID uint64) *EventMessage {
	return &EventMessage{
		AreaID:    areaID,
		Param:     make(map[string]interface{}),
		EventType: et,
	}
}

type HandleFunc func(em EventMessage) error

func (e *EventMessage) SetDeviceID(deviceID int) {
	e.Param["device_id"] = deviceID
}

func (e *EventMessage) GetDeviceID() int {
	if v, ok := e.Param["device_id"].(int); ok {
		return v
	}
	return 0
}

func (e *EventMessage) SetAttr(attr entity.Attribute) {
	e.Param["attr"] = attr
}

func (e *EventMessage) GetAttr() *entity.Attribute {
	if v, ok := e.Param["attr"]; ok {
		var attr entity.Attribute
		bytes, _ := json.Marshal(v)
		json.Unmarshal(bytes, &attr)
		return &attr
	}
	return nil
}
