package resource

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type restartReq struct {
	Id   string `uri:"id"`
	Name string `json:"name" form:"name"`
}

// 不能重启服务
var unRestartList = map[string]struct{}{
	"smartassistant": {},
	"zt-nginx":       {},
	"zt-vue":         {},
}

func RestartContainer(c *gin.Context) {
	var (
		err  error
		req  restartReq
		resp interface{}
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindUri(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if _, flag := unRestartList[req.Name]; flag == true {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	dClient := docker.GetClient().DockerClient
	inspect, err := dClient.ContainerInspect(c, req.Id)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if inspect.Config.Labels == nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if req.Name != inspect.Config.Labels["com.docker.compose.service"] {
		if req.Name != inspect.Config.Labels["com.zhiting.smartassistant.resource.service_name"] {
			err = errors.Wrap(err, errors.BadRequest)
			return
		}
	}

	if err = dClient.ContainerRestart(c, req.Id, nil); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

}
