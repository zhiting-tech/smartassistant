package task

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/device"

	"github.com/google/uuid"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// TaskFunc 任务运行函数
type TaskFunc func(task *Task) error

// WrapperFunc 运行 TaskFunc 的 Wrapper
type WrapperFunc func(TaskFunc) TaskFunc

// Task 将场景转化为任务
type Task struct {
	ID       string
	Value    string // The Value of the item; arbitrary.
	Priority int64  // 使用
	// The index is needed by update and is maintained by the heap.Interface methods.
	index    int // The index of the item in the heap.
	f        TaskFunc
	Parent   *Task // 父任务
	wrappers []WrapperFunc
}

// NewTaskAt 按运行时间点创建任务
func NewTaskAt(f TaskFunc, t time.Time) *Task {
	return &Task{
		ID:       uuid.New().String(),
		Value:    "",
		Priority: t.Unix(),
		f:        f,
	}
}

// NewTask 按延迟运行时间创建任务
func NewTask(f TaskFunc, delay time.Duration) *Task {
	return NewTaskAt(f, time.Now().Add(delay))
}

// WithParent 设置父任务
func (item *Task) WithParent(parent *Task) *Task {
	item.Parent = parent
	return item
}

// WithWrapper 设置 Wrapper
func (item *Task) WithWrapper(wrappers ...WrapperFunc) *Task {
	item.wrappers = append(item.wrappers, wrappers...)
	return item
}

// Run 执行
// TODO
func (item *Task) Run() {
	logger.Debug("Run ", item.ToString())
	if item.f != nil {
		f := item.f
		for _, wrapper := range item.wrappers {
			f = wrapper(f)
		}
		if err := f(item); err != nil {
			logger.Error("task run err:", err)
		}
	}
}

func (item *Task) ToString() string {
	return fmt.Sprintf("Task Value %s, Priority %d, index %d", item.Value, item.Priority, item.index)
}

// taskLogWrapper 包装任务执行函数，在任务执行时插入日志，任务执行完后更新日志
func taskLogWrapper(target interface{}) WrapperFunc {
	return func(f TaskFunc) TaskFunc {
		return func(task *Task) error {
			if task == nil {
				return nil
			}
			var parentID *string
			if task.Parent != nil {
				parentID = &task.Parent.ID
			}
			// if s, ok := target.(entity.Scene); ok { // 记录场景正在执行的task的index
			//	MinHeapQueue.SceneTaskIndexMap.Store(s.ID, task.index)
			// }
			if err := entity.NewTaskLog(target, task.ID, parentID); err != nil {
				logger.Error("NewTaskLogErr:", err)
			}
			err := f(task)
			if e := entity.UpdateTaskLog(task.ID, err); e != nil {
				logger.Error(e)
			}
			return err
		}
	}
}

// IsConditionsSatisfied 场景条件是否满足 isTrigByTimer 是否由定时条件触发
func IsConditionsSatisfied(scene entity.Scene, isTrigByTimer bool) bool {
	if !scene.IsOn {
		logger.Debugf("scene %d: is off\n", scene.ID)
		return false
	}
	if !IsInTimePeriod(scene) { // 不在有效时间段内则不执行
		logger.Debugf("scene %d: not in effective time period\n", scene.ID)
		return false
	}
	// “任一满足”情况下，定时触发的任务直接满足条件
	if !scene.IsMatchAllCondition() && isTrigByTimer {
		return true
	}
	for _, condition := range scene.SceneConditions {
		if condition.ConditionType == entity.ConditionTypeTiming {
			continue
		}

		// 任一满足
		if !scene.IsMatchAllCondition() && IsConditionSatisfied(condition) {
			logger.Debugf("scene %d: condition:%d satisfied\n", scene.ID, condition.ID)
			return true
		}
		// 全部满足（有一个不满足）
		if scene.IsMatchAllCondition() && !IsConditionSatisfied(condition) {
			logger.Debugf("scene %d: condition:%d not satisfied\n", scene.ID, condition.ID)
			return false
		}
	}

	logger.Debugf("scene.ID %d, scene.ConditionLogic %d \n", scene.ID, scene.ConditionLogic)
	return scene.IsMatchAllCondition()
}

// IsInTimePeriod 是否在时间段内
func IsInTimePeriod(scene entity.Scene) bool {

	weekday := time.Now().Weekday()
	if !strings.Contains(scene.RepeatDate, strconv.Itoa(int(weekday))) {
		logger.Debugf("scene %d: today not in repeat date\n", scene.ID)
		return false
	}

	if scene.TimePeriodType == entity.TimePeriodTypeCustom {
		days := int(time.Now().Sub(scene.EffectStart).Hours() / 24)
		effectEndTime := scene.EffectEnd.AddDate(0, 0, days)
		effectStartTime := scene.EffectStart.AddDate(0, 0, days)
		return time.Now().Before(effectEndTime) && time.Now().After(effectStartTime)
	}
	return true
}

// IsConditionSatisfied 判断设备状态是否满足条件
func IsConditionSatisfied(condition entity.SceneCondition) bool {
	if condition.ConditionType == entity.ConditionTypeTiming {
		return false
	}

	d, err := entity.GetDeviceByID(condition.DeviceID)
	if err != nil {
		logger.Errorf("get device %d err:%v\n", condition.DeviceID, err)
		return false
	}

	var item entity.Attribute
	if err = json.Unmarshal(condition.ConditionAttr, &item); err != nil {
		logger.Error("Unmarshal error:", err)
		return false
	}
	shadow, err := device.GetShadow(d)
	if err != nil {
		logger.Error("GetAttribute error:", err)
		return false
	}
	val, err := shadow.Get(item.InstanceID, item.Attribute.Attribute)
	if err != nil {
		logger.Error("shadow get attribute error:", err)
		return false
	}
	logger.Debugf("%v %s %v\n", val, condition.Operator, item.Val)
	return item.Operate(condition.Operator, val)
}
