package supervisor

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/clouddisk"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/disk"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/filebrowser"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"os"
	"path/filepath"
	"strings"
)

type PathType int
type fileType int

const (
	FirstPath   PathType = iota
	BackupPath           // backup
	Internal             // 内部存储
	External             // 外部存储
	UnmountDisk          // 未挂载硬盘
	Pool                 // 存储池
	Partition            // 存储分区
	Resource             // 内部存储文件夹或文件
	MountedRes           // 外部已挂载硬盘下的文件夹或文件
)

const (
	FolderTypeDir fileType = iota
	FolderTypeFile

	InternalName = "内部存储"
	ExternalName = "外部存储"
)

type BackupPathsReq struct {
	Path    string   `form:"path"`
	Type    PathType `form:"type"`
	OnlyDir bool     `form:"only_dir"`
}

type BackupPathsResp struct {
	Paths []backupPathInfo `json:"paths"`
}

type backupPathInfo struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Type     PathType `json:"type"`
	FileType fileType `json:"file_type"`
}

func ListBackupPath(c *gin.Context) {
	var (
		err  error
		req  BackupPathsReq
		resp BackupPathsResp
	)

	defer func() {
		if len(resp.Paths) == 0 {
			resp.Paths = make([]backupPathInfo, 0)
		}
		response.HandleResponse(c, err, resp)
	}()

	accessToken := c.GetHeader(types.SATokenKey)
	ctx := c.Request.Context()

	if err = c.BindQuery(&req); err != nil {
		return
	}

	switch req.Type {
	case FirstPath:
		if req.Path == "" {
			resp.Paths = handlerFirstPath(ctx)
		}
	case Internal:
		if req.Path == "/" {
			resp.Paths = handlerInternal(accessToken, ctx)
		}
	case Pool:
		resp.Paths = handlerPool(accessToken, req.Path, ctx)
	case Partition:
		resp.Paths = handlerPartition(accessToken, req.Path, ctx)
	case Resource:
		resp.Paths = handlerResource(accessToken, req, ctx)
	case MountedRes, BackupPath:
		if req.Type == BackupPath && req.OnlyDir {
			return
		}
		if req.Type == MountedRes {
			req.Path = filepath.Join("/volume", req.Path)
		}
		if resp.Paths, err = handlerMountedDiskFolders(req.Path, req.OnlyDir, req.Type); err != nil {
			return
		}
	case External:
		resp.Paths = handlerExternal(ctx)
	}
}

// 处理一级目录信息
func handlerFirstPath(ctx context.Context) (resp []backupPathInfo) {
	paths := []backupPathInfo{{
		Name: InternalName,
		Path: "/",
		Type: Internal,
	}, {
		Name: ExternalName,
		Path: "/",
		Type: External,
	}}
	resp = append(resp, paths...)
	client, err := disk.NewDiskManagerClient()
	if err != nil {
		logger.Warnf("handlerFirstPath new DiskManager Client err is %s", err)
		return
	}
	result, err := client.GetPhysicalVolumeListWithContext(ctx)
	if err != nil {
		logger.Warnf("handlerFirstPath get physical err is %s", err)
		return
	}
	for _, r := range result.PVS {
		if r.VGName == "" && !r.IsMounted {
			resp = append(resp, backupPathInfo{
				Name: r.Name,
				Type: UnmountDisk,
			})
		}
	}
	return
}

// handlerInternal 处理内部存储信息
func handlerInternal(accessToken string, ctx context.Context) (resp []backupPathInfo) {
	resp = append(resp, backupPathInfo{
		Name: "backup",
		Path: fmt.Sprintf("%c%s", os.PathSeparator, "backup"),
		Type: BackupPath,
	})
	poolsResp, err := clouddisk.GetCloudDiskPools(accessToken, ctx)
	if err != nil {
		logger.Warnf("handlerInternal get pools err is %s", err)
		return
	}
	for _, p := range poolsResp.Data.Pools {
		resp = append(resp, backupPathInfo{
			Name: p.Name,
			Path: filepath.Join("/", p.Name),
			Type: Pool,
		})
	}
	return
}

// handlerExternal 处理外部存储信息
func handlerExternal(ctx context.Context) (resp []backupPathInfo) {
	client, err := disk.NewDiskManagerClient()
	if err != nil {
		logger.Warnf("handlerFirstPath new DiskManager Client err is %s", err)
		return
	}
	result, err := client.GetPhysicalMountedListWithContext(ctx)
	if err != nil {
		logger.Warnf("handlerMountedPhysical get physical mounted err is %s", err)
		return
	}
	for _, r := range result.PVS {
		resp = append(resp, backupPathInfo{
			Name: r.Name,
			Path: filepath.Base(r.Name),
			Type: MountedRes,
		})
	}
	return
}

// handlerMountedDiskFolders 处理已挂载的外部硬盘下的文件列表
func handlerMountedDiskFolders(path string, onlyDir bool, pathType PathType) (resp []backupPathInfo, err error) {
	fs := filebrowser.GetFBOrInit()
	isDir, err := fs.IsDir(path)
	if err != nil {
		return
	}
	if !isDir {
		err = errors.Wrap(err, status.FileNotExistErr)
		return
	}

	file, err := fs.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Wrap(err, status.FileNotExistErr)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}
	defer file.Close()
	fileInfos, err := file.Readdir(-1)
	if err != nil {
		return
	}

	for _, f := range fileInfos {
		ft := FolderTypeDir
		if !f.IsDir() {
			ft = FolderTypeFile
		}
		if onlyDir && !f.IsDir() {
			continue
		}
		if !onlyDir && !f.IsDir() && filepath.Ext(f.Name()) != ".zip" {
			continue
		}
		resp = append(resp, backupPathInfo{
			Name:     f.Name(),
			Path:     filepath.Join(path, f.Name()),
			Type:     pathType,
			FileType: ft,
		})
	}
	return
}

// handlerPool 处理内部存储池信息
func handlerPool(accessToken, poolName string, ctx context.Context) (resp []backupPathInfo) {
	poolName = strings.TrimPrefix(poolName, "/")
	poolInfo, err := clouddisk.GetCloudDiskPoolInfo(accessToken, poolName, ctx)
	if err != nil {
		logger.Warnf("handlerPool get pool info err is %s", err)
		return
	}
	for _, l := range poolInfo.Data.Lv {
		resp = append(resp, backupPathInfo{
			Name: l.Name,
			Path: filepath.Join("/", poolName, l.Name),
			Type: Partition,
		})
	}
	return
}

// handlerPartition 处理内部存储分区信息
func handlerPartition(accessToken, path string, ctx context.Context) (resp []backupPathInfo) {
	partitionInfo, err := clouddisk.GetCloudDiskPartitionInfo(accessToken, path, ctx)
	if err != nil {
		logger.Warnf("handlerPartition get partition info err is %s", err)
		return
	}
	for _, p := range partitionInfo.Data.FolderInfo {
		resp = append(resp, backupPathInfo{
			Name: p.Name,
			Path: p.Path,
			Type: Resource,
		})
	}
	return
}

// handlerResource 处理内部存储文件资源信息
func handlerResource(accessToken string, req BackupPathsReq, ctx context.Context) (resp []backupPathInfo) {
	folderInfo, err := clouddisk.GetCloudDiskFolderInfo(accessToken, req.Path, ctx)
	if err != nil {
		logger.Warnf("handlerResource get folder info err is %s", err)
		return
	}
	for _, f := range folderInfo.Data.FolderList {
		if req.OnlyDir && f.Type == int(FolderTypeFile) {
			continue
		}
		if !req.OnlyDir && f.Type == int(FolderTypeFile) && filepath.Ext(f.Name) != ".zip" {
			continue
		}
		resp = append(resp, backupPathInfo{
			Name:     f.Name,
			Path:     f.Path,
			Type:     Resource,
			FileType: fileType(f.Type),
		})
	}
	return
}
