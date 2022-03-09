package docker

import (
	"bufio"
	"context"
	"encoding/json"
	errors2 "errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/hashicorp/go-version"
	jsoniter "github.com/json-iterator/go"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// Image 镜像
type Image struct {
	Name     string `json:"name" yaml:"name"`
	Tag      string `json:"tag" yaml:"tag"`           // 镜像标签
	Registry string `json:"registry" yaml:"registry"` // 仓库地址
}

func (i Image) RefStr() string {
	tag := i.Tag
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("%s:%s", i.Repository(), tag)
}

func (i Image) Repository() string {
	if i.Registry == "" {
		return i.Name
	}
	return fmt.Sprintf("%s/%s", i.Registry, i.Name)
}

// IsImageNewest TODO 镜像是否最新
func (c *Client) IsImageNewest() (isNewest bool, err error) {
	return false, nil
}

func (c *Client) GetImageNewestTag(img Image) (tag string, err error) {
	var tagResp struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://%v/v2/%v/tags/list", img.Registry, img.Name),
		nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = jsoniter.Unmarshal(data, &tagResp); err != nil {
		return
	}
	versions := make([]*version.Version, 0, len(tagResp.Tags))
	for _, t := range tagResp.Tags {
		if sv, e := version.NewSemver(t); e == nil {
			versions = append(versions, sv)
		}
	}
	if len(versions) == 0 {
		err = errors.New("no version")
		return
	}
	sort.Sort(sort.Reverse(version.Collection(versions)))
	return versions[0].String(), nil
}

// IsImageAdd 镜像是否已经拉取到本地
func (c *Client) IsImageAdd(refStr string) (isAdded bool) {
	ctx := context.Background()
	inspect, b, err := c.DockerClient.ImageInspectWithRaw(ctx, refStr)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Debug(inspect)
	logger.Debug(string(b))
	return true
}

// Pull 拉取镜像
func (c *Client) Pull(refStr string) (err error) {
	logger.Info("pulling image: ", refStr)
	ctx := context.Background()
	readCloser, err := c.DockerClient.ImagePull(ctx, refStr, types.ImagePullOptions{})
	if err != nil {
		logger.Warning(err)
		return
	}
	defer readCloser.Close()
	ioutil.ReadAll(readCloser)
	logger.Info("pull success")
	return
}

// ImageRemove 删除镜像
func (c *Client) ImageRemove(refStr string) (err error) {
	_, err = c.DockerClient.ImageRemove(context.Background(), refStr, types.ImageRemoveOptions{Force: true})
	if err != nil {
		return
	}
	logger.Debug("image removed")
	return
}

// ImageSave save docker images to tar file.
func (c *Client) ImageSave(target string, imgs ...string) (err error) {
	ids := make([]string, 0, len(imgs))
	for _, img := range imgs {
		ids = append(ids, img)
	}
	readCloser, err := c.DockerClient.ImageSave(context.Background(), ids)
	if err != nil {
		return
	}
	defer readCloser.Close()
	outFile, err := os.Create(target)
	if err != nil {
		return
	}
	_, err = io.Copy(outFile, readCloser)
	return
}

// ImageLoad load docker image(s) from a tar file.
func (c *Client) ImageLoad(target string) (err error) {
	f, err := os.Open(target)
	if err != nil {
		return
	}
	defer f.Close()
	res, err := c.DockerClient.ImageLoad(context.Background(), f, true)
	if err != nil {
		return
	}
	res.Body.Close()
	return
}
func (c *Client) Inspect(imageID string) (inspect types.ImageInspect, err error) {
	inspect, _, err = c.DockerClient.ImageInspectWithRaw(context.Background(), imageID)
	if err != nil {
		return
	}
	return
}

func (c *Client) BuildFromPath(path, tag string) (imageID string, err error) {

	tar, err := archive.TarWithOptions(path, &archive.TarOptions{})
	if err != nil {
		return
	}
	defer tar.Close()
	return c.BuildFromTar(tar, tag)
}

func (c *Client) BuildFromTar(tar io.Reader, tag string) (imageID string, err error) {

	opts := types.ImageBuildOptions{Remove: true, Tags: []string{tag}}
	resp, err := c.DockerClient.ImageBuild(context.Background(), tar, opts)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var res LineResult
		json.Unmarshal(scanner.Bytes(), &res)
		if res.Error != "" {
			logrus.Error("err: ", res.Error, res.ErrorDetail.Message)
			return "", errors2.New(res.Error)
		}
		if strings.HasPrefix(res.Stream, "Successfully built") {
			imageID = strings.TrimRight(strings.TrimPrefix(res.Stream, "Successfully built "), "\n")
			logrus.Debug("image id:", imageID)
		}
	}

	return
}

type LineResult struct {
	Stream string `json:"stream"`
	AUX    struct {
		ID string `json:"id"`
	} `json:"aux"`
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}
type ErrorDetail struct {
	Message string `json:"message"`
}
