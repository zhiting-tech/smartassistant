package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

// Lock 智能门锁
type Lock struct {
	LockTargetState *LockTargetState
	Logs			*Logs
	Battery 		*Battery
}

// LockTargetState 锁的目标状态
type LockTargetState struct {
	attribute.Int
}

func NewLockTargetState() *LockTargetState {
	return &LockTargetState{}
}

// Logs 日志
type Logs struct {
	attribute.String
}

func NewLogs() *Logs {
	return &Logs{}
}