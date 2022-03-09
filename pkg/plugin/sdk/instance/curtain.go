package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

// Curtain 窗帘
type Curtain struct {
	CurrentPosition *Position `tag:"required"` // 当前位置 0-100 TODO 不可写字段
	TargetPosition  *Position `tag:"required"` // 目标位置 0-100
	State           *State    `tag:"required"` // 0关1开2暂停
	Style           *Style    // 0左右1左开2右开3上下

	// TODO 考虑窗帘和窗帘控制器分开定义
	Direction  *Direction // 0默认方向1反方向
	UpperLimit *Limit     // 0删除1设置
	LowerLimit *Limit     // 0删除1设置
	Calibration  *Calibration  // 校准 true为进行校准/已经校准完 false为未校准
}

func (c Curtain) InstanceName() string {
	return "curtain"
}

// Position 窗帘行程位置
type Position struct {
	attribute.Int
}

func NewPosition() *Position {
	p := Position{}
	p.SetRange(0, 100)
	return &p
}

// State 窗帘状态
type State struct {
	attribute.Enum
}

func NewState() *State {
	a := State{}
	a.SetEnums(0, 1, 2) // close/open/pause
	return &a
}

// Direction 方向
type Direction struct {
	attribute.Enum
}

func NewDirection() *Direction {
	d := Direction{}
	d.SetEnums(0, 1) // 默认方向/反方向
	return &d
}

// Limit 上/下限
type Limit struct {
	attribute.Enum
}

func NewLimit() *Limit {
	l := Limit{}
	l.SetEnums(0, 1) // 删除/设置上限
	return &l
}

// Style 窗帘样式
type Style struct {
	attribute.Enum
}

func NewStyle() *Style {
	s := Style{}
	s.SetEnums(0, 1, 2, 3) // 左右/左/右/上下
	return &s
}

type Calibration struct {
	attribute.Bool
}

func NewCalibration() *Calibration {
	return &Calibration{}
}