package scene

import (
	"time"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/task"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// CreateSceneReq 创建场景接口请求参数
type CreateSceneReq struct {
	SceneInfo
}

// SceneInfo 新场景的配置信息
type SceneInfo struct {
	entity.Scene
	SceneConditions []entity.ConditionInfo `json:"scene_conditions"`
	CreateTime      int64                  `json:"create_time"`
	EffectStartTime int64                  `json:"effect_start_time"`
	EffectEndTime   int64                  `json:"effect_end_time"`
}

// CreateScene 用于处理创建场景接口的请求
func CreateScene(c *gin.Context) {
	var (
		req CreateSceneReq
		err error
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		return
	}

	if err = req.check(c); err != nil {
		return
	}
	if err = req.createScene(c); err != nil {
		return
	}

	// 手动场景创建时不排进任务队列
	if req.AutoRun {
		task.GetManager().AddSceneTask(req.Scene)
	}
}

func (req *CreateSceneReq) check(c *gin.Context) (err error) {
	if !entity.JudgePermit(session.Get(c).UserID, types.SceneAdd) {
		err = errors.New(status.SceneCreateDeny)
		return
	}

	if err = req.validate(c); err != nil {
		return
	}
	return
}

func (req *CreateSceneReq) validate(c *gin.Context) (err error) {

	if len(req.SceneTasks) == 0 {
		err = errors.New(errors.BadRequest)
		return
	}

	if err = entity.IsSceneNameExist(req.Name, req.ID, session.Get(c).AreaID); err != nil {
		return
	}

	// 手动执行
	if !req.AutoRun {
		if req.TimePeriodType != 0 && req.ConditionLogic != 0 && req.RepeatType != 0 &&
			req.EffectStartTime != 0 && req.EffectEndTime != 0 && req.RepeatDate != "" && len(req.SceneConditions) != 0 {
			err = errors.New(errors.BadRequest)
			return
		}
	} else {
		// 自动执行
		if len(req.SceneConditions) == 0 {
			err = errors.Newf(status.SceneParamIncorrectErr, "触发条件")
			return
		}
		// 生效时间不可为空
		if req.EffectStartTime == 0 || req.EffectEndTime == 0 {
			err = errors.Newf(status.SceneParamIncorrectErr, "生效时间")
			return
		}

		// ConditionLogic 校验
		if req.CheckConditionLogic() {
			err = errors.Newf(status.SceneParamIncorrectErr, "满足条件类型")
			return
		}
		// 生效时间类型检验
		if err = req.CheckPeriodType(); err != nil {
			return
		}

		// SceneCondition 触发条件检验
		var count int
		for _, sc := range req.SceneConditions {
			// 触发条件为满足全部时，定时触发条件只允许一个
			if sc.ConditionType == entity.ConditionTypeTiming && req.IsMatchAllCondition() {
				count++
				if count > 1 {
					err = errors.New(status.ConditionTimingCountErr)
					return
				}

			}
			if err = sc.CheckCondition(session.Get(c).UserID); err != nil {
				return
			}
		}
	}
	// 执行任务的校验
	for _, sceneTask := range req.SceneTasks {
		if err = CheckSceneTasks(c, sceneTask); err != nil {
			return
		}
	}
	return
}

// CheckSceneTasks 执行任务校验
func CheckSceneTasks(c *gin.Context, task entity.SceneTask) (err error) {
	userId := session.Get(c).UserID
	if err = task.CheckTaskType(); err != nil {
		err = errors.New(status.TaskTypeErr)
		return
	}
	// 控制设备
	if task.Type == entity.TaskTypeSmartDevice {
		if err = task.CheckTaskDevice(userId); err != nil {
			return
		}
	} else {
		if err = checkTaskScene(c, task.ControlSceneID); err != nil {
			return
		}
	}
	return
}

// checkTaskScene 校验场景任务类型
func checkTaskScene(c *gin.Context, controlSceneId int) (err error) {
	// 控制场景
	if err = entity.CheckSceneExitById(controlSceneId); err != nil {
		return
	}

	var controlPermission bool
	if controlPermission, err = CheckControlPermission(c, controlSceneId, session.Get(c).UserID); err != nil {
		return
	}
	if !controlPermission {
		err = errors.New(status.DeviceOrSceneControlDeny)
		return
	}

	return
}

func (req *CreateSceneReq) createScene(c *gin.Context) (err error) {
	u := session.Get(c)
	req.Scene.CreatorID = u.UserID
	req.Scene.EffectStart = time.Unix(req.EffectStartTime, 0)
	req.Scene.EffectEnd = time.Unix(req.EffectEndTime, 0)
	// 自动场景
	if req.AutoRun {
		req.Scene.SceneConditions = getConditionReq(req.SceneConditions)
		req.Scene.IsOn = true
	}

	req.Scene.SceneTasks = req.SceneTasks
	// 添加场景所属家庭
	req.Scene.AreaID = u.AreaID
	if err = entity.CreateScene(&req.Scene); err != nil {
		return
	}
	return
}

// getConditionReq 获取触发条件参数值
func getConditionReq(conditions []entity.ConditionInfo) (sceneConditions []entity.SceneCondition) {
	for _, condition := range conditions {
		condition.TimingAt = time.Unix(condition.Timing, 0)
		sceneConditions = append(sceneConditions, condition.SceneCondition)
	}
	return
}
