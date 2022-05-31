package supervisor

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"go.opentelemetry.io/otel/trace"

	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/modules/supervisor/proto"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	errors2 "github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

const (
	stageDir = "stage"
)

var (
	manager *Manager
	_once   sync.Once
)

func currentSAImage() docker.Image {
	return docker.Image{
		Name:     "zhitingtech/smartassistant",
		Tag:      types.Version,
		Registry: config.GetConf().SmartAssistant.DockerRegistry,
	}
}

type Manager struct {
	// 运行时目录，docker-compose.yaml 所在
	RuntimePath string
	BackupPath  string
}

func GetManager() *Manager {
	_once.Do(func() {
		ab, err := filepath.Abs(config.GetConf().SmartAssistant.BackupPath())
		if err != nil {
			logger.Errorf("backup path not valid: %v", err)
		}
		ar, err := filepath.Abs(config.GetConf().SmartAssistant.RuntimePath)
		if err != nil {
			logger.Errorf("runtime path not valid: %v", err)
		}
		manager = &Manager{
			RuntimePath: ar,
			BackupPath:  ab,
		}
		_ = os.MkdirAll(manager.BackupPath, os.ModePerm)
		f, err := os.Stat(manager.BackupPath)
		if os.IsNotExist(err) {
			logger.Errorf("can not create backup path %v", manager.BackupPath)
		}
		if !f.IsDir() {
			logger.Error("backup path is not a dir")
		}
	})
	return manager
}

func (m *Manager) ListBackups() []Backup {
	bks := make([]Backup, 0)
	filepath.Walk(m.BackupPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// TODO 只查找第一层的zip文件
		if info.IsDir() {
			return nil
		}
		if filepath.Dir(path) != m.BackupPath {
			return nil
		}
		if filepath.Ext(path) != ".zip" {
			return nil
		}
		bks = append(bks, NewBackupFromFileName(info.Name()))
		return nil
	})
	return bks
}

func stopAllPlugins() (err error) {
	resumeContainer := func(plgs []plugin.Plugin) {
		for _, plg := range plgs {
			_, _ = plugin.RunPlugin(plg)
		}
	}
	plgs, err := plugin.GetGlobalManager().LoadPluginsWithContext(context.TODO())
	cli := docker.GetClient()
	if err != nil {
		return
	}
	stoppedPlgs := make([]plugin.Plugin, 0)
	for _, plg := range plgs {
		ps, _ := docker.GetClient().ContainerIsRunningByImage(plg.Image)
		if ps == false {
			continue
		}

		err = cli.StopContainer(plg.Image)
		if err != nil {
			resumeContainer(stoppedPlgs)
			return err
		} else {
			stoppedPlgs = append(stoppedPlgs, *plg)
		}
	}
	return
}

func startAllPlugins() (err error) {
	plgs, err := plugin.GetGlobalManager().LoadPluginsWithContext(context.TODO())
	if err != nil {
		return
	}
	for _, plg := range plgs {
		plugin.RunPlugin(*plg)
	}
	return
}

func (m *Manager) processRestart(cn string) (err error) {
	err = stopAllPlugins()
	if err != nil {
		return
	}
	// 返回操作结果后重启
	go func() {
		time.Sleep(time.Second)
		err = docker.GetClient().DockerClient.ContainerRestart(context.Background(),
			cn, nil)
		if err != nil {
			logger.Warnf("restart self error %v", err)
		}
	}()
	return
}

// StartBackupJob 停止所有插件,通知supervisor备份
func (m *Manager) StartBackupJobWithContext(ctx context.Context, req entity.BackupInfo) (err error) {
	err = stopAllPlugins()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			nerr := startAllPlugins()
			if nerr != nil {
				logger.Warnf("Supervisor start all plugins error %v", nerr)
			}
		}
	}()
	newCtx := trace.ContextWithSpan(context.Background(), trace.SpanFromContext(ctx))
	go func() {
		time.Sleep(time.Second)
		err = GetClient().BackupSmartassistantWithContext(newCtx, req)
		if err != nil {
			logger.Errorf("backup error %v", err)
		}
	}()

	return
}

// StartRestoreJob 停止所有插件,通知supervisor还原数据
func (m *Manager) StartRestoreJobWithContext(ctx context.Context, file string) (err error) {
	err = stopAllPlugins()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			nerr := startAllPlugins()
			if nerr != nil {
				logger.Warnf("Supervisor start all plugins error %v", nerr)
			}
		}
	}()
	newCtx := trace.ContextWithSpan(context.Background(), trace.SpanFromContext(ctx))
	go func() {
		time.Sleep(time.Second)
		err = GetClient().RestoreSmartassistantWithContext(newCtx, file)
		if err != nil {
			logger.Errorf("restore error %v", err)
		}
	}()

	return
}

func (m *Manager) DeleteBackup(fn string) error {
	fn = filepath.Join(m.BackupPath, filepath.Clean("/"+fn))
	fi, err := os.Stat(fn)
	if err != nil {
		if os.IsNotExist(err) {
			return errors2.New(status.FileNotExistErr)
		}
		return err
	}
	if fi.IsDir() {
		return errors2.New(status.FileNotExistErr)
	}
	return os.RemoveAll(fn)
}

// StartUpdateJob 下载新版镜像，通知supervisor以新镜像重启
func (m *Manager) StartUpdateJobWithContext(ctx context.Context, version string) (err error) {
	var (
		result *cloud.SoftwareLastVersionHttpResult
		req    proto.UpdateReq
	)
	if result, err = cloud.GetLastSoftwareVersionWithContext(ctx); err != nil {
		return
	}

	req.SoftwareVersion = result.Data.Version
	for _, subservice := range result.Data.Services {
		if !docker.GetClient().IsImageAdd(subservice.Image) {
			logger.Debugf("image not found, pull %s", subservice.Image)
			err = docker.GetClient().Pull(subservice.Image)
			if err != nil {
				err = errors2.New(status.ImagePullErr)
				return
			}
		}
		req.UpdateItems = append(req.UpdateItems, &proto.UpdateItem{
			ServiceName: subservice.Name,
			NewImage:    subservice.Image,
			Version:     subservice.Version,
		})
	}

	err = stopAllPlugins()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			nerr := startAllPlugins()
			if nerr != nil {
				logger.Warnf("Supervisor start all plugins error %v", nerr)
			}
		}
	}()

	err = GetClient().UpdateWithContext(ctx, &req)
	if err != nil {
		code := grpcStatus.Code(err)
		if code == codes.Unavailable {
			err = errors2.Wrap(err, status.SupervisorNotStart)
		} else {
			err = errors2.Wrap(err, errors2.InternalServerErr)
		}
	}
	return
}
