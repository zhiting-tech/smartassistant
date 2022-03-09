package utils

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/instance"
)

type testDevice struct {
	LightBulb instance.LightBulb
	Info      instance.Info
}

func newTestDevice() *testDevice {
	return &testDevice{
		LightBulb: instance.LightBulb{
			Power: attribute.NewPower(),
		},
		Info: instance.Info{
			Name:         &attribute.Name{},
			Identity:     &attribute.Identity{},
			Model:        &attribute.Model{},
			Manufacturer: &attribute.Manufacturer{},
			Version:      &attribute.Version{},
		},
	}
}

func TestCustomDevice(t *testing.T) {

	logrus.SetLevel(logrus.DebugLevel)
	td := newTestDevice()
	d := Parse(td)
	assert.Equal(t, 2, len(d.Instances))
	assert.Equal(t, 3, len(d.Instances[0].Attributes))
	assert.Equal(t, 5, len(d.Instances[1].Attributes))

	type CustomAttr struct {
		attribute.Base
	}
	type CustomLightBulb struct {
		instance.LightBulb
		Custom0 *CustomAttr
		Custom1 *CustomAttr
	}
	type CustomDevice struct {
		LightBulb *CustomLightBulb
	}

	custom := &CustomDevice{LightBulb: &CustomLightBulb{
		LightBulb: instance.LightBulb{
			Power:      &attribute.Power{},
			ColorTemp:  &instance.ColorTemp{},
			Brightness: &instance.Brightness{},
		},
		Custom0: &CustomAttr{},
		Custom1: nil,
	}}
	device := Parse(custom)
	assert.Equal(t, 1, len(device.Instances))
	assert.Equal(t, 5, len(device.Instances[0].Attributes))
	for i, attr := range device.Instances[0].Attributes {
		assert.Equal(t, i+1, attr.ID, attr.Name)
		if attr.Model == nil {
			assert.False(t, attr.Active, attr.Name)
		} else {
			assert.True(t, attr.Active, attr.Name)
		}
	}
}

func TestParse(t *testing.T) {
	d := Parse(nil)
	assert.Nil(t, d)

	s := struct{}{}
	d = Parse(s)
	assert.Nil(t, d.Instances)

	s2 := struct {
		I  int
		ss string
		A  attribute.Name
		b  attribute.Name
		l  instance.LightBulb
	}{}
	d = Parse(s2)
	assert.Nil(t, d.Instances)

	s3 := struct {
		L    instance.LightBulb
		Info instance.Info
	}{}

	s3.Info = instance.Info{
		Name:         attribute.NewName(),
		Identity:     attribute.NewIdentity(),
		Model:        attribute.NewModel(),
		Manufacturer: attribute.NewManufacturer(),
		Version:      attribute.NewVersion(),
	}
	s3.Info.Name.SetString("testname")
	d = Parse(s3)
	assert.Equal(t, 1, len(d.Instances))

	s4 := struct {
		Child instance.ChildInstances
	}{}
	s4.Child.Instances = append(s4.Child.Instances, instance.LightBulb{}, instance.ColorTemp{})
	g := Parse(s4)

	fmt.Println(g.GetAttribute(1, "power"))
}

type customIns struct {
	A attribute.Name
	a attribute.Name
	B *attribute.Name
}

func (a customIns) InstanceName() string {
	return "custom"
}
func TestDefineAttr(t *testing.T) {

	logrus.SetLevel(logrus.DebugLevel)
	s := struct {
		L customIns
	}{L: customIns{
		A: attribute.Name{},
		a: attribute.Name{},
		B: &attribute.Name{},
	}}
	d := Parse(s)
	assert.Equal(t, 1, len(d.Instances))
	assert.Equal(t, 1, len(d.Instances[0].Attributes))
	as := d.Instances[0].Attributes[0].Model.(attribute.StringType)
	val := "hello"
	as.SetString(val)
	assert.Equal(t, val, s.L.B.GetString())
}

func TestUpdateAttr(t *testing.T) {
	td := newTestDevice()
	d := Parse(td)
	attr := d.Instances[0].Attributes[0].Model.(attribute.StringType)
	val := "toggle"
	attr.SetString(val)
	assert.Equal(t, val, td.LightBulb.Power.GetString())
}

// BenchmarkParse
// BenchmarkParse-12            	   38437	     31399 ns/op	    5276 B/op	     227 allocs/op
func BenchmarkParse(b *testing.B) {
	td := newTestDevice()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(td)
	}
}

// BenchmarkParseParallel
// BenchmarkParseParallel-12    	  159999	      6644 ns/op	    5551 B/op	     227 allocs/op
func BenchmarkParseParallel(b *testing.B) {
	td := newTestDevice()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Parse(td)
		}
	})

}
