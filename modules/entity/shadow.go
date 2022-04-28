package entity

import (
	"errors"
)

type State struct {
	// Desired 期望值
	Desired map[string]map[int]interface{} `json:"desired"`
	// Reported 报告值
	Reported map[string]map[int]interface{} `json:"reported"`
}

type Metadata struct {
	Desired  map[string]map[int]AttrMetadata `json:"desired"`
	Reported map[string]map[int]AttrMetadata `json:"reported"`
}

type AttrMetadata struct {
	Timestamp int64 `json:"timestamp"`
}

// Shadow shadow of device
type Shadow struct {
	State     State    `json:"state"`
	Metadata  Metadata `json:"metadata"`
	Timestamp int      `json:"timestamp"`
	Version   int      `json:"version"`
}

func NewShadow() Shadow {
	return Shadow{
		State: State{
			Desired:  make(map[string]map[int]interface{}),
			Reported: make(map[string]map[int]interface{}),
		},
		Metadata: Metadata{
			Desired:  make(map[string]map[int]AttrMetadata),
			Reported: make(map[string]map[int]AttrMetadata),
		},
	}
}

func (s *Shadow) UpdateReported(iid string, aid int, val interface{}) {

	if ins, ok := s.State.Reported[iid]; ok {
		ins[aid] = val
	} else {
		s.State.Reported[iid] = map[int]interface{}{aid: val}
	}
}

func (s Shadow) reportedAttr(iid string, aid int) (val interface{}, err error) {
	if ins, ok := s.State.Reported[iid]; ok {

		if val, ok = ins[aid]; ok {
			return val, nil
		}
	}
	err = errors.New("attr not found in shadow")
	return

}

func (s Shadow) Get(iid string, aid int) (val interface{}, err error) {
	return s.reportedAttr(iid, aid)
}
