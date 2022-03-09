package archive

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	zipFile := "test.zip"

	dir, err := os.Getwd()
	assert.NoError(t, err, nil)

	// 创建测试压缩文件
	zipDir, err := ioutil.TempDir(dir, "temp")
	assert.NoError(t, err, nil)
	defer os.RemoveAll(zipDir)
	f1, err := ioutil.TempFile(zipDir, "temp")
	assert.NoError(t, err, nil)
	f1.Close()
	tempDir2, err := ioutil.TempDir(zipDir, "temp")
	assert.NoError(t, err, nil)
	f2, err := ioutil.TempFile(tempDir2, "temp")
	assert.NoError(t, err, nil)
	f2.Close()

	err = Zip(zipFile, zipDir)
	assert.NoError(t, err, nil)
	defer os.Remove(zipFile)

	tempDir, err := ioutil.TempDir(dir, "temp")
	assert.NoError(t, err, nil)
	defer os.RemoveAll(tempDir)

	err = UnZip(tempDir, zipFile)
	assert.NoError(t, err, nil)

	// 匹配文件夹和文件
	fileMap := make(map[string]struct{})
	filepath.Walk(tempDir, func(path string, info fs.FileInfo, err error) error {
		key := strings.TrimPrefix(path, filepath.Dir(path)+string(filepath.Separator))
		fileMap[key] = struct{}{}
		return nil
	})

	filepath.Walk(zipDir, func(path string, info fs.FileInfo, err error) error {
		key := strings.TrimPrefix(path, filepath.Dir(path)+string(filepath.Separator))
		_, ok := fileMap[key]
		assert.Equal(t, ok, true)
		return nil
	})

}
