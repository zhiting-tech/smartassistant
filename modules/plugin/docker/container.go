package docker

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/regex"
)

// ContainerIsRunningByImage 返回是否有镜像对应的容器在运行
func (c *Client) ContainerIsRunningByImage(image string) (isRunning bool, err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.DockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}
	for _, con := range containers {
		if con.Image == image {
			return true, nil
		}
	}
	return false, nil
}

// ContainerRun 根据镜像创建容器并运行
func (c *Client) ContainerRun(image string, conf container.Config, hostConf container.HostConfig) (containerID string, err error) {
	ctx := context.Background()

	logger.Info("create container ", image)
	r, err := c.DockerClient.ContainerCreate(ctx, &conf, &hostConf,
		nil, nil, regex.ToSnakeCase(image))
	if err != nil {
		logger.Error("ContainerCreateErr", err)
		return
	}
	logger.Info("start container ", r.ID)
	containerID = r.ID
	err = c.DockerClient.ContainerStart(ctx, r.ID, types.ContainerStartOptions{})
	if err != nil {
		logger.Error("ContainerStart", err)
		return
	}
	return
}

// StopContainer 停止容器 TODO 优化
func (c *Client) StopContainer(image string) (err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.DockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, con := range containers {
		if con.Image == image {
			logger.Debugf("stop container %s:%s", con.ID, image)
			err = c.DockerClient.ContainerStop(ctx, con.ID, nil)
			if err != nil {
				return
			}
			logger.Debugf("container %s:%s stop", con.ID, image)
			return
		}
	}
	return
}

// RemoveContainer 删除容器 TODO 优化
func (c *Client) RemoveContainer(image string) (err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.DockerClient.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return
	}

	for _, con := range containers {
		if con.Image == image {
			logger.Debugf("remove container %s:%s", con.ID, image)
			err = c.DockerClient.ContainerRemove(ctx, con.ID, types.ContainerRemoveOptions{})
			if err != nil {
				return
			}
			logger.Debugf("container %s:%s removed", con.ID, image)
			return
		}
	}
	return
}

// ContainerRestartByImage 重启容器
func (c *Client) ContainerRestartByImage(image string) (err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.DockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, con := range containers {
		if con.Image == image {
			logger.Debugf("restart container %s:%s", con.ID, image)
			err = c.DockerClient.ContainerRestart(ctx, con.ID, nil)
			if err != nil {
				return
			}
			logger.Debugf("container %s:%s restarted", con.ID, image)
		}
	}
	return nil
}

func (c *Client) GetContainerByImage(image string) (id string, err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.DockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, con := range containers {
		logger.Infof("container %v, %v", con.ID, con.Image)
		if con.Image == image {
			id = con.ID
			return
		}
	}
	return "", errors.New("not found")
}

// ContainerKillByImage 给容器发送信号
func (c *Client) ContainerKillByImage(image string, signal string) (err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.DockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, con := range containers {
		if con.Image == image {
			logger.Debugf("kill container %s", image)
			err = c.DockerClient.ContainerKill(ctx, con.ID, signal)
			if err != nil {
				return
			}
			logger.Debugf("container %s killed", con.ImageID)
		}
	}
	return nil
}

// ContainerList 返回启动的容器列表
func (c *Client) ContainerList() ([]types.Container, error) {
	ctx := context.Background()
	return c.DockerClient.ContainerList(ctx, types.ContainerListOptions{})
}
