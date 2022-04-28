package definer

import "github.com/zhiting-tech/smartassistant/pkg/thingmodel"

type AttributeEvent struct {
	IID string      `json:"iid"`
	AID int         `json:"aid"`
	Val interface{} `json:"val"`
}

type ThingModelEvent struct {
	ThingModel thingmodel.ThingModel `json:"thingModel"`
	IID        string
}
