package event

import (
	"encoding/json"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
)

type EventType string

const (
	AttributeChange  EventType = "attribute_change"
	DeviceIncrease   EventType = "device_increase"
	DeviceDecrease   EventType = "device_decrease"
	ThingModelChange EventType = "thing_model_change"
	OnlineStatus     EventType = "online_status"
)

type EventMessage struct {
	ID        int64
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

func (e *EventMessage) SetDeviceID(d int) {
	e.Param["device_id"] = d
}

func (e *EventMessage) GetDeviceID() int {
	if v, ok := e.Param["device_id"].(int); ok {
		return v
	}
	return 0
}

func (e *EventMessage) SetAttr(attr definer.AttributeEvent) {
	e.Param["attr"] = attr
}

func (e *EventMessage) GetAttr() *definer.AttributeEvent {
	if v, ok := e.Param["attr"]; ok {
		var attr definer.AttributeEvent
		bytes, _ := json.Marshal(v)
		json.Unmarshal(bytes, &attr)
		return &attr
	}
	return nil
}
