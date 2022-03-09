package attribute

type FloatType interface {
	GetRange() (*float64, *float64)
	SetRange(float64, float64)
	GetFloat() float64
	SetFloat(float64)
}

type Float struct {
	Base
	min, max *float64
	v        float64
}

func (f *Float) SetRange(min, max float64) {
	if min > max {
		return
	}
	f.min = &min
	f.max = &max
}

func (f *Float) GetRange() (min, max *float64) {
	return f.min, f.max
}

func (f *Float) SetFloat(v float64) {
	f.v = v
}

func (f *Float) GetFloat() float64 {
	return f.v
}
