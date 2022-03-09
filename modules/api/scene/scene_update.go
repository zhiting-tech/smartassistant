package scene

import (
	"strconv"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/task"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// UpdateSceneReq 修改场景接口请求参数
type UpdateSceneReq struct {
	DelConditionIds []int `json:"del_condition_ids"`
	DelTaskIds      []int `json:"del_task_ids"`
	CreateSceneReq
}

// UpdateScene 用于处理修改场景接口的请求
func UpdateScene(c *gin.Context) {
	var (
		req UpdateSceneReq
		err error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	sceneId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	if err = c.BindJSON(&req); err != nil {
		return

	}

	if err = req.validateRequest(sceneId, c); err != nil {
		return
	}

	req.wrapReq()

	if err = req.updateScene(sceneId); err != nil {
		return
	}

	if e := task.GetManager().RestartSceneTask(sceneId); e != nil {
		logger.Error("restart scene task err:", e)
	}

}

func (req *UpdateSceneReq) validateRequest(sceneId int, c *gin.Context) (err error) {
	scene, err := entity.GetSceneById(sceneId)
	if err != nil {
		return
	}
	// 场景类型不允许修改
	if scene.AutoRun != req.AutoRun {
		err = errors.New(status.SceneTypeForbidModify)
		return
	}

	if err = req.CreateSceneReq.validate(c); err != nil {
		return
	}
	return
}

func (req UpdateSceneReq) updateScene(sceneId int) (err error) {

	if err = entity.GetDB().Session(&gorm.Session{FullSaveAssociations: true}).Transaction(func(tx *gorm.DB) error {
		return entity.UpdateSceneByIDWithTx(sceneId, &req.Scene, tx)
	}); err != nil {
		return
	}

	// TODO 后续调整为使用事务
	if err = req.delConditions(sceneId); err != nil {
		return
	}

	if err = req.delTasks(sceneId); err != nil {
		return
	}

	return
}

func (req *UpdateSceneReq) wrapReq() {
	req.Scene.CreatedAt = time.Unix(req.CreateTime, 0)
	req.Scene.EffectStart = time.Unix(req.EffectStartTime, 0)
	req.Scene.EffectEnd = time.Unix(req.EffectEndTime, 0)

	for _, sc := range req.SceneConditions {
		sceneCondition := sc.SceneCondition
		sceneCondition.TimingAt = time.Unix(sc.Timing, 0)

		req.Scene.SceneConditions = append(req.Scene.SceneConditions, sceneCondition)
	}
}

// delConditions 删除触发条件
func (req *UpdateSceneReq) delConditions(sceneId int) (err error) {
	if len(req.DelConditionIds) == 0 {
		return
	}

	var conditions []entity.SceneCondition
	for _, id := range req.DelConditionIds {
		condition := entity.SceneCondition{ID: id, SceneID: sceneId}
		conditions = append(conditions, condition)
	}
	if err = entity.GetDB().Delete(conditions).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

// delTasks 删除执行任务
func (req *UpdateSceneReq) delTasks(sceneId int) (err error) {
	if len(req.DelTaskIds) == 0 {
		return
	}

	var tasks []entity.SceneTask
	for _, id := range req.DelTaskIds {
		task := entity.SceneTask{ID: id, SceneID: sceneId}
		tasks = append(tasks, task)
	}
	if err = entity.GetDB().Delete(tasks).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}
