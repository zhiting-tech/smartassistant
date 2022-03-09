package docker

import (
	"sync"

	"github.com/docker/docker/client"
)

var (
	defaultClient *Client
	clientOnce    sync.Once
)

type Client struct {
	DockerClient *client.Client
}

func GetClient() *Client {
	clientOnce.Do(func() {
		dockerClient, err := client.NewClientWithOpts(client.FromEnv,
			client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		defaultClient = &Client{
			DockerClient: dockerClient,
		}
	})
	return defaultClient
}
