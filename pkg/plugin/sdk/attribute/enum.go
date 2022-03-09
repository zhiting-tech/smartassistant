package attribute

type EnumType interface {
	SetEnums(enums ...int)
	GetEnum() int
	SetEnum(enum int)
}

type Enum struct {
	Base
	v     int
	enums map[int]struct{}
}

func (e *Enum) SetEnums(enums ...int) {
	if e.enums == nil {
		e.enums = make(map[int]struct{})
	}
	for i := range enums {
		e.enums[i] = struct{}{}
	}
}

func (e *Enum) GetEnum() int {
	return e.v
}

func (e *Enum) SetEnum(enum int) {
	e.v = enum
}
