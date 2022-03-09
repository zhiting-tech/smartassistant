package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

type LightBulb struct {
	Power      *attribute.Power `tag:"required"`
	ColorTemp  *ColorTemp
	Brightness *Brightness
	Hue        *Hue
	Saturation *Saturation
	RGB        *RGB
}

func (l LightBulb) InstanceName() string {
	return "light_bulb"
}

type ColorTemp struct {
	attribute.Int
}

func NewColorTemp() *ColorTemp {
	return &ColorTemp{}
}

type Brightness struct {
	attribute.Int
}

func NewBrightness() *Brightness {
	return &Brightness{}
}

type Hue struct {
	attribute.Int
}

func NewHue() *Hue {
	return &Hue{}
}

type Saturation struct {
	attribute.Int
}

func NewSaturation() *Saturation {
	return &Saturation{}
}

type RGB struct {
	attribute.String
}

func NewRGB() *RGB {
	return &RGB{}
}
