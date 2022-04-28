package filebrowser

import (
	"github.com/spf13/afero"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
)


// File represents a file in the filesystem.
type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	io.WriterAt

	Readdir(count int) ([]os.FileInfo, error)
}

// FileBrowser 接管业务对文件/文件夹相关的操作
type FileBrowser struct {
	fs       afero.Fs    // 底层文件操作封装
	root     string      // 根目录
	fileMode os.FileMode // 创建文件的默认权限
	dirMode  os.FileMode // 创建目录的默认权限
}

func (fb *FileBrowser) GetRoot() string {
	return fb.root
}

func (fb *FileBrowser) Create(path string) (File, error) {
	return fb.fs.Create(path)
}

// Remove 删除文件
func (fb *FileBrowser) Remove(path string) error {
	return fb.fs.Remove(path)
}

// RemoveAll 删除目录以及所有文件
func (fb *FileBrowser) RemoveAll(path string) error {
	return fb.fs.RemoveAll(path)
}

// Mkdir 会递归创建目录
func (fb *FileBrowser) Mkdir(path string) error {
	return fb.fs.MkdirAll(path, fb.dirMode)
}

func (fb *FileBrowser) Open(path string) (File, error) {
	return fb.fs.Open(path)
}

func (fb *FileBrowser) stat(path string) (os.FileInfo, error) {
	return fb.fs.Stat(path)
}

func (fb *FileBrowser) Stat(path string) (fileInfo os.FileInfo, err error) {
	fileInfo, err = fb.stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Wrap(err, status.FileNotExistErr)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

func (fb *FileBrowser) Rename(oldPath, name string) error {
	err := fb.fs.Rename(oldPath, name)
	logger.Errorf("fs rename err %v", err)
	return err
}

func (fb *FileBrowser) IsDir(path string) (isDir bool, err error) {
	fileInfo, err := fb.fs.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Wrap(err, status.FileNotExistErr)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}

	isDir = fileInfo.IsDir()
	return
}

func (fb *FileBrowser) CopyFile(source, dest string) (err error) {
	// Open the source file.
	src, err := fb.fs.Open(source)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	defer src.Close()

	// 新建目标文件
	_, srcName := filepath.Split(source)
	dest = filepath.Join(dest, srcName)

	dst, err := fb.fs.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
	if err != nil {
		return err
	}
	defer dst.Close()

	// 复制源文件的内容到目标文件
	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	// 给目标文件赋予源文件的权限
	info, err := fb.fs.Stat(source)
	if err != nil {
		return err
	}
	err = fb.fs.Chmod(dest, info.Mode())
	if err != nil {
		return err
	}

	return nil
}

func (fb *FileBrowser) CopyDir(source, dest string) (err error) {

	// 获取源文件信息
	srcInfo, err := fb.fs.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Wrap(err, status.FileNotExistErr)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}

	_, srcName := filepath.Split(source)
	dest = filepath.Join(dest, srcName)

	// 创建目标目录
	err = fb.fs.MkdirAll(dest, srcInfo.Mode())
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return err
	}

	srcFile, _ := fb.fs.Open(source)

	files, err := srcFile.Readdir(-1)
	if err != nil {
		errors.Wrap(err, errors.InternalServerErr)
		return
	}

	for _, file := range files {
		fSource := source + "/" + file.Name()
		// 是目录则递归调用
		if file.IsDir() {
			err = fb.CopyDir(fSource, dest)
			if err != nil {
				return
			}
		} else {
			err = fb.CopyFile(fSource, dest)
			if err != nil {
				return
			}
		}
	}

	return
}

func (fb *FileBrowser) MoveFile(source, dest string) (err error) {

	_, fileSuffix := filepath.Split(source)
	targetPath := filepath.Join(dest, fileSuffix)
	if err = fb.fs.Rename(source, targetPath); err != nil {
		if os.IsNotExist(err) {
			err = errors.Wrap(err, status.FileNotExistErr)
			return
		}
		logger.Errorf("fs MoveFile err %v", err)
		err = errors.Wrap(err, errors.InternalServerErr)
		return err
	}
	return
}

var fb *FileBrowser
var once sync.Once

func GetFBOrInit() *FileBrowser {
	once.Do(func() {
		fb = &FileBrowser{
			fs:       nil,
			dirMode:  0777,
			fileMode: 0666,
		}
		rootPath := config.GetConf().SmartAssistant.RuntimePath
		if !path.IsAbs(rootPath) {
			wd, err := os.Getwd()
			if err != nil {
				log.Fatalf("can not read current dir, error: %v", err.Error())
			}
			rootPath = filepath.Join(wd, rootPath)
		}
		log.Printf("use %v as file root path", rootPath)

		if err := os.MkdirAll(rootPath, fb.dirMode); err != nil {
			log.Fatalf("can not create root data dir, error: %v", err.Error())
		}
		fb.root = rootPath
		fb.fs = afero.NewBasePathFs(afero.NewOsFs(), rootPath)

	})
	return fb
}
