package attribute

type IntType interface {
	GetRange() (*int, *int)
	SetRange(int, int)
	GetInt() int
	SetInt(int)
}

type Int struct {
	Base
	min, max *int
	v        int
}

func (i *Int) SetRange(min, max int) {
	if min > max {
		return
	}
	i.min = &min
	i.max = &max
}

func (i *Int) GetRange() (min, max *int) {
	return i.min, i.max
}

func (i *Int) SetInt(v int) {
	i.v = v
}

func (i *Int) GetInt() int {
	return i.v
}

