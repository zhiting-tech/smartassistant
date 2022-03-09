package scene

import (
	"sort"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// 场景日志接口返回日志的默认数量
const logSizeDefault = 40

// ListSceneTaskReq 场景日志接口请求参数
type ListSceneTaskReq struct {
	Start int `form:"start"`
	Size  int `form:"size"`
}

// ListSceneTaskLogResp 场景日志接口返回数据
type ListSceneTaskLogResp []DateLogInfo

// DateLogInfo 某月的场景日志
type DateLogInfo struct {
	Date              string             `json:"date"`
	SceneTaskLogInfos []SceneTaskLogInfo `json:"items"`
}

// SceneTaskLogInfo 场景日志信息
type SceneTaskLogInfo struct {
	Name       string                `json:"name"`
	Type       entity.TaskType       `json:"type"`
	Result     entity.TaskResultType `json:"result"`
	FinishedAt int64                 `json:"finished_at"`
	Items      []TaskLogItem         `json:"items"`
}

// TaskLogItem 场景执行任务信息
type TaskLogItem struct {
	Name         string                `json:"name"`
	Type         entity.TaskType       `json:"type"`
	LocationName string                `json:"location_name,omitempty"`
	DepartmentName string				`json:"department_name,omitempty"`
	Result       entity.TaskResultType `json:"result"`
}

// ListSceneTaskLog 用于处理场景日志接口的请求
func ListSceneTaskLog(c *gin.Context) {
	var (
		err error
		req ListSceneTaskReq

		taskLogs []entity.TaskLog
		resp     ListSceneTaskLogResp
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.BindQuery(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if req.Size == 0 {
		req.Size = logSizeDefault
	}

	if err = entity.GetDB().
		Preload("ChildTaskLogs").
		Order("finished_at desc").
		Where("type !=? and finish=? and result !=? and area_id=?",
			entity.TaskTypeSmartDevice, true, entity.TaskSceneAlreadyDeleted, session.Get(c).AreaID).
		Offset(req.Start).
		Limit(req.Size).
		Find(&taskLogs).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if resp, err = LogInfosGroupByDate(taskLogs); err != nil {
		return
	}

	return
}

func LogInfosGroupByDate(taskLogs []entity.TaskLog) (logInfos []DateLogInfo, err error) {
	var (
		logDates     []string
		dateLogInfos = make(map[string][]SceneTaskLogInfo)
	)

	logInfos = make([]DateLogInfo, 0)

	for _, taskLog := range taskLogs {
		if !taskLog.Finish {
			continue
		}

		taskLogInfo := SceneTaskLogInfo{
			Name:       taskLog.Name,
			Type:       taskLog.Type,
			Result:     taskLog.Result,
			FinishedAt: taskLog.FinishedAt.Unix(),
			Items:      WrapLogItems(taskLog),
		}
		date := taskLog.FinishedAt.Format("2006-01")
		if _, ok := dateLogInfos[date]; !ok {
			dateLogInfos[date] = make([]SceneTaskLogInfo, 0)
		}
		dateLogInfos[date] = append(dateLogInfos[date], taskLogInfo)
	}

	for d, _ := range dateLogInfos {
		logDates = append(logDates, d)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(logDates)))
	for _, d := range logDates {
		logInfos = append(logInfos, DateLogInfo{
			Date:              d,
			SceneTaskLogInfos: dateLogInfos[d],
		})
	}
	return
}

func WrapLogItems(taskLog entity.TaskLog) (taskItems []TaskLogItem) {

	taskItems = make([]TaskLogItem, 0)

	// 任务部分执行成功/执行失败时展示执行详情
	if taskLog.Result == entity.TaskPartSuccess || taskLog.Result == entity.TaskFail && len(taskLog.ChildTaskLogs) != 0 {
		for _, taskLog := range taskLog.ChildTaskLogs {
			taskItems = append(taskItems, TaskLogItem{
				Name:         taskLog.Name,
				Type:         taskLog.Type,
				Result:       taskLog.Result,
				LocationName: taskLog.DeviceLocation,
				DepartmentName: taskLog.DeviceDepartment,
			})
		}
	}

	return
}
