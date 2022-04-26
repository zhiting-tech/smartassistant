package scene

import (
	errors2 "errors"
	"net/http"
	"strconv"

	device2 "github.com/zhiting-tech/smartassistant/modules/device"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// InfoSceneResp 场景详情接口返回数据
type InfoSceneResp struct {
	entity.Scene
	CreateTime      int64           `json:"create_time"`
	EffectStartTime int64           `json:"effect_start_time"`
	EffectEndTime   int64           `json:"effect_end_time"`
	SceneConditions []ConditionInfo `json:"scene_conditions"`
	SceneTasks      []SceneTaskInfo `json:"scene_tasks"`
	AreaIDStr       string          `json:"area_id"`
}

// ConditionInfo 场景触发条件信息
type ConditionInfo struct {
	entity.ConditionInfo
	DeviceInfo `json:"device_info"`
}

// SceneTaskInfo 场景执行任务信息
type SceneTaskInfo struct {
	entity.SceneTask
	ControlSceneInfo ControlSceneInfo `json:"control_scene_info"`
	DeviceInfo       `json:"device_info"`
}

// ControlSceneInfo 执行任务类型为场景时,任务场景信息
type ControlSceneInfo struct {
	Name   string      `json:"name"`
	Status sceneStatus `json:"status"`
}

// DeviceInfo 执行任务类型为设备时,任务设备信息
type DeviceInfo struct {
	Name           string       `json:"name"`
	LocationName   string       `json:"location_name,omitempty"`
	DepartmentName string       `json:"department_name,omitempty"`
	LogoURL        string       `json:"logo_url"`
	Status         deviceStatus `json:"status"`
}

// InfoScene 用于处理场景详情接口的请求
func InfoScene(c *gin.Context) {
	var (
		err     error
		resp    InfoSceneResp
		sceneID int
		scene   entity.Scene
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if sceneID, err = strconv.Atoi(c.Param("id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	scene, err = entity.GetSceneInfoById(sceneID)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.SceneNotExist)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if err = WrapResponse(c, scene, &resp); err != nil {
		return
	}

	return
}

func WrapResponse(c *gin.Context, scene entity.Scene, resp *InfoSceneResp) (err error) {

	var (
		conditionInfo   ConditionInfo
		sceneTask       SceneTaskInfo
		sceneConditions = make([]ConditionInfo, 0)
		sceneTasks      = make([]SceneTaskInfo, 0)
	)

	for _, condition := range scene.SceneConditions {

		if conditionInfo, err = WrapConditionInfo(c, condition); err != nil {
			return
		}

		sceneConditions = append(sceneConditions, conditionInfo)
	}

	for _, task := range scene.SceneTasks {
		if sceneTask, err = WrapTaskInfo(c, task); err != nil {
			return
		}
		sceneTasks = append(sceneTasks, sceneTask)
	}

	resp.SceneConditions = sceneConditions
	resp.Scene = scene
	resp.SceneTasks = sceneTasks
	resp.CreateTime = scene.CreatedAt.Unix()
	resp.EffectStartTime = scene.EffectStart.Unix()
	resp.EffectEndTime = scene.EffectEnd.Unix()
	resp.AreaIDStr = strconv.FormatUint(scene.AreaID, 10)

	return
}

func WrapConditionInfo(c *gin.Context, condition entity.SceneCondition) (conditionInfo ConditionInfo, err error) {
	var (
		deviceInfo DeviceInfo
	)

	if condition.ConditionType == entity.ConditionTypeDeviceStatus {
		if deviceInfo, err = WrapDeviceInfo(condition.DeviceID, c.Request, c); err != nil {
			return
		}
	}

	conditionInfo.Timing = condition.TimingAt.Unix()
	conditionInfo.SceneCondition = condition
	conditionInfo.DeviceInfo = deviceInfo

	return
}

func WrapTaskInfo(c *gin.Context, task entity.SceneTask) (taskInfo SceneTaskInfo, err error) {
	var (
		scene entity.Scene
	)
	taskInfo = SceneTaskInfo{
		SceneTask: task,
	}

	if task.Type != entity.TaskTypeSmartDevice {
		if scene, err = entity.GetSceneByIDWithUnscoped(task.ControlSceneID); err != nil {
			if errors2.Is(err, gorm.ErrRecordNotFound) {
				err = errors.Wrap(err, status.SceneNotExist)
				return
			}
			err = errors.Wrap(err, errors.InternalServerErr)
			return
		}
		taskInfo.ControlSceneInfo.Name = scene.Name
		// 场景已被删除
		if scene.Deleted.Valid {
			taskInfo.ControlSceneInfo.Status = sceneAlreadyDelete
			return
		}
		taskInfo.ControlSceneInfo.Status = sceneNormal
		return
	}

	if taskInfo.DeviceInfo, err = WrapDeviceInfo(taskInfo.DeviceID, c.Request, c); err != nil {
		return
	}

	return

}

func WrapDeviceInfo(deviceID int, req *http.Request, c *gin.Context) (deviceInfo DeviceInfo, err error) {
	var (
		location   entity.Location
		department entity.Department
		device     entity.Device
	)

	if device, err = entity.GetDeviceByIDWithUnscoped(deviceID); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.DeviceNotExist)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if device.LocationID != 0 {
		if location, err = entity.GetLocationByID(device.LocationID); err != nil {
			return
		}
	}
	if device.DepartmentID != 0 {
		if department, err = entity.GetDepartmentByID(device.DepartmentID); err != nil {
			return
		}
	}

	deviceInfo.Name = device.Name
	deviceInfo.LocationName = location.Name
	deviceInfo.DepartmentName = department.Name
	deviceInfo.LogoURL = device2.LogoURL(req, device)

	if device.Deleted.Valid {
		// 设备已删除
		deviceInfo.Status = deviceAlreadyDelete
		return
	}

	// TODO:判断设备是否可连接
	deviceInfo.Status = deviceNormal
	return

}
