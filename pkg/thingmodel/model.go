package thingmodel

import (
	"errors"
	"fmt"
)

type ThingModel struct {
	Instances []Instance `json:"instances"`

	// OTASupport 是否支持OTA
	OTASupport bool `json:"ota_support"`
}

func (das ThingModel) GetAttribute(iid string, aid int) (Attribute, error) {

	ins, err := das.GetInstance(iid)
	if err != nil {
		return Attribute{}, err
	}
	return ins.GetAttribute(aid)
}

func (das ThingModel) GetInstance(iid string) (Instance, error) {

	for _, ins := range das.Instances {
		if ins.IID == iid {
			return ins, nil
		}
	}
	return Instance{}, fmt.Errorf("instance %s not found", iid)
}

// GetInfo 获取设备基础信息
func (das ThingModel) GetInfo(iid string) (info DeviceInfo, err error) {

	ins, err := das.GetInstance(iid)
	if err != nil {
		return
	}
	return ins.GetInfo()
}

func (das ThingModel) IsGateway() bool {
	for _, i := range das.Instances {
		if i.IsGateway() {
			return true
		}
	}
	return false
}

type DeviceInfo struct {
	IID          string
	Name         string
	Model        string
	Manufacturer string
	Version      string
}

type Instance struct {
	IID      string    `json:"iid"`
	Services []Service `json:"services"`
}

func (i Instance) GetInfo() (info DeviceInfo, err error) {

	for _, srv := range i.Services {
		if srv.Type != "info" {
			continue
		}
		for _, attr := range srv.Attributes {
			switch attr.Type {
			case Name.String():
				info.Name, _ = attr.Val.(string)
			case Model.String():
				info.Model, _ = attr.Val.(string)
			case Manufacturer.String():
				info.Manufacturer, _ = attr.Val.(string)
			case Identify.String():
				info.IID, _ = attr.Val.(string)
			case Version.String():
				info.Version, _ = attr.Val.(string)
			}
		}

		if info.Name == "" {
			info.Name = info.Model
		}
		return
	}
	err = errors.New("info service not found")
	return
}

func (i Instance) IsGateway() bool {
	for _, s := range i.Services {
		if s.Type == GatewayService {
			return true
		}
	}
	return false
}

func (i Instance) GetAttribute(aid int) (Attribute, error) {
	for _, s := range i.Services {
		for _, a := range s.Attributes {
			if a.AID == aid {
				return a, nil
			}
		}
	}
	return Attribute{}, fmt.Errorf("attribute %d not found", aid)
}
