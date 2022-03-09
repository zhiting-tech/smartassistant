package attribute

import "github.com/sirupsen/logrus"

type StringType interface {
	GetString() string
	SetString(string)
}

type String struct {
	Base
	v           string
	validValues map[string]interface{}
}

func (s *String) SetString(v string) {
	if len(s.validValues) != 0 {
		if _, ok := s.validValues[v]; !ok {
			logrus.Warning("invalid string value: ", v)
			// TODO return error
			return
		}
	}
	s.v = v
}

func (s String) GetString() string {
	return s.v
}