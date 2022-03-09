package attribute

type BoolType interface {
	GetBool() bool
	SetBool(bool)
}

type Bool struct {
	Base
	v bool
}

func (b *Bool) SetBool(v bool) {
	b.v = v
}

func (b *Bool) GetBool() bool {
	return b.v
}