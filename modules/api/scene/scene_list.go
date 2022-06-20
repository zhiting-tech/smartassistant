package scene

import (
	errors2 "errors"
	"net/http"

	device2 "github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/plugin"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// 场景状态
type sceneStatus int

// 场景状态:1,正常;2,已删除;
const (
	sceneNormal sceneStatus = iota + 1
	sceneAlreadyDelete
)

// 设备状态
type deviceStatus int

// 设备状态:1,正常;2,已删除;3,离线;
const (
	deviceNormal deviceStatus = iota + 1
	deviceAlreadyDelete
	deviceDisConnect
)

// 场景列表过滤条件
type listType int

// 0:所有场景;1:有权限的场景
const (
	allScene listType = iota
	permitScene
)

// sceneListReq 场景列表接口请求参数
type sceneListReq struct {
	Type listType `form:"type"`
}

// sceneListResp 场景列表接口返回数据
type sceneListResp struct {
	Manual  []manualSceneInfo  `json:"manual"`
	AutoRun []autoRunSceneInfo `json:"auto_run"`
}

// Scene 场景基础信息
type Scene struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	ControlPermission bool   `json:"control_permission"`
}

// manualSceneInfo 手动执行的场景信息
type manualSceneInfo struct {
	Scene
	Items []Item `json:"items"`
}

// autoRunSceneInfo 自动执行的场景信息
type autoRunSceneInfo struct {
	Scene
	IsOn      bool           `json:"is_on"`
	Condition sceneCondition `json:"condition"`
	Items     []Item         `json:"items"`
}

// sceneCondition 场景触发条件信息
type sceneCondition struct {
	Type    entity.ConditionType `json:"type"`
	LogoURL string               `json:"logo_url"`
	Status  int                  `json:"status"`
}

// Item 场景执行任务信息
type Item struct {
	ID      int             `json:"-"`
	Type    entity.TaskType `json:"type"`
	LogoURL string          `json:"logo_url"`
	Status  int             `json:"status"`
	devices []entity.Attribute
}

// ListScene 用于处理场景列表接口的请求
func ListScene(c *gin.Context) {
	var (
		err    error
		req    sceneListReq
		resp   sceneListResp
		scenes []entity.Scene
		user   *session.User
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	user = session.Get(c)

	if err = c.BindQuery(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if scenes, err = entity.GetScenes(user.AreaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if resp.Manual, resp.AutoRun, err = WrapScenes(c, scenes, user.UserID, req.Type); err != nil {
		return
	}

	return
}

func WrapScenes(c *gin.Context, scenes []entity.Scene, userID int, listType listType) (manualScenes []manualSceneInfo, autoRunScenes []autoRunSceneInfo, err error) {
	var (
		items             []Item
		condition         sceneCondition
		controlPermission bool
	)

	manualScenes = make([]manualSceneInfo, 0)
	autoRunScenes = make([]autoRunSceneInfo, 0)

	for _, scene := range scenes {

		if controlPermission, err = CheckControlPermission(c, scene.ID, userID); err != nil {
			return
		}
		if listType == permitScene && !controlPermission {
			continue
		}
		// 场景信息
		sceneInfo := Scene{
			ID:                scene.ID,
			Name:              scene.Name,
			ControlPermission: controlPermission,
		}

		// 场景执行任务信息
		if items, err = WrapTasks(c, scene.ID); err != nil {
			return
		}

		// 场景触发条件信息()
		if scene.AutoRun {
			// 自动触发条件
			if condition, err = WrapCondition(c, scene.ID); err != nil {
				return
			}

			autoRunScene := autoRunSceneInfo{
				Scene:     sceneInfo,
				Items:     items,
				Condition: condition,
				IsOn:      scene.IsOn,
			}
			autoRunScene.Condition = condition
			autoRunScenes = append(autoRunScenes, autoRunScene)

		} else {
			// 手动没有触发条件
			manualScene := manualSceneInfo{
				Scene: sceneInfo,
				Items: items,
			}
			manualScenes = append(manualScenes, manualScene)
		}

	}
	return
}

func WrapCondition(ctx *gin.Context, sceneID int) (sceneCondition sceneCondition, err error) {
	var (
		conditions []entity.SceneCondition
	)

	// 获取场景的所有触发条件
	if conditions, err = entity.GetConditionsBySceneID(sceneID); err != nil {
		return
	}

	for i, c := range conditions {
		// 只返回第一个触发条件的信息
		sceneCondition.Type = conditions[0].ConditionType
		if c.ConditionType == entity.ConditionTypeDeviceStatus {
			// 第一个触发条件为设备时，包装对应信息
			if i == 0 {
				item := Item{ID: c.DeviceID}
				if err = WrapDeviceItem(&item, ctx.Request); err != nil {
					return
				}
				sceneCondition.LogoURL = item.LogoURL
				sceneCondition.Status = item.Status
			}
		}
	}
	return
}

func WrapTasks(c *gin.Context, sceneID int) (items []Item, err error) {
	var (
		tasks []entity.SceneTask
		item  Item
	)

	items = make([]Item, 0)
	// 获取场景所有执行任务
	if tasks, err = entity.GetSceneTasksBySceneID(sceneID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	for _, task := range tasks {
		if item, err = WrapTask(c, task); err != nil {
			return
		}
		items = append(items, item)
	}
	return
}

func WrapTask(c *gin.Context, task entity.SceneTask) (item Item, err error) {
	var (
		taskDevices []entity.Attribute
		scene       entity.Scene
	)

	item.Type = task.Type

	// 执行任务类型为智能设备
	if task.Type == entity.TaskTypeSmartDevice {
		item.ID = task.DeviceID
		item.devices = taskDevices
		if err = WrapDeviceItem(&item, c.Request); err != nil {
			return
		}
		return
	}
	// 执行任务类型为场景
	item.ID = task.ControlSceneID
	if scene, err = entity.GetSceneByIDWithUnscoped(task.ControlSceneID); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.SceneNotExist)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if scene.Deleted.Valid {
		item.Status = int(sceneAlreadyDelete)
		return
	}
	item.Status = int(sceneNormal)
	return
}

func CheckControlPermission(c *gin.Context, sceneID int, userID int) (controlPermission bool, err error) {
	checked := make(map[int]bool)
	return checkControlPermission(c, sceneID, userID, checked)
}

func checkControlPermission(c *gin.Context, sceneID int, userID int, checked map[int]bool) (controlPermission bool, err error) {
	var (
		items []Item
	)
	if v, ok := checked[sceneID]; ok {
		controlPermission = v
		return
	}

	// 没有控制场景的权限，直接返回
	if !entity.JudgePermit(userID, types.SceneControl) {
		return
	}
	controlPermission = true
	checked[sceneID] = true

	if items, err = WrapTasks(c, sceneID); err != nil {
		return
	}
	var up entity.UserPermissions
	up, err = entity.GetUserPermissions(userID)
	if err != nil {
		return
	}
	for _, item := range items {
		// 校验执行任务为智能设备时对该设备的控制权限
		if item.Type == entity.TaskTypeSmartDevice {

			// 已删除的设备跳过判断
			if item.Status == int(deviceAlreadyDelete) {
				continue
			}
			// 判断设备每一个操作的控制权限
			for _, device := range item.devices {
				if !up.IsDeviceAttrControlPermit(item.ID, device.AID) {
					controlPermission = false
					checked[sceneID] = false
					return
				}
			}
			continue
		}

		if controlPermission, err = checkControlPermission(c, item.ID, userID, checked); err != nil {
			return
		}
		// 嵌套控制场景不满足权限就直接返回false
		if !controlPermission {
			return
		}
	}
	return
}

func WrapDeviceItem(item *Item, req *http.Request) (err error) {
	var (
		device entity.Device
	)

	if device, err = entity.GetDeviceByIDWithUnscoped(item.ID); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.DeviceNotExist)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	item.LogoURL = device2.LogoURL(req, device)

	if device.Deleted.Valid {
		// 设备已删除
		item.Status = int(deviceAlreadyDelete)
		return
	}
	if plugin.GetGlobalClient().IsOnline(plugin.Identify{
		PluginID: device.PluginID,
		IID:      device.IID,
		AreaID:   device.AreaID,
	}) {
		item.Status = int(deviceNormal)
	} else {
		item.Status = int(deviceDisConnect)
	}

	return
}
