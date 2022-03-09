package cloud

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	errors2 "errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/setting"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/archive"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gorm.io/gorm"
)

const (
	HttpRequestTimeout = (time.Duration(30) * time.Second)
)

func getBaseTempDir() string {
	path := path.Join(config.GetConf().SmartAssistant.RuntimePath, "run", "smartassistant", "temp")
	_, err := os.Stat(path)
	if err != nil {
		os.MkdirAll(path, os.ModePerm)
	}
	return path
}

type AreaMigrationReq struct {
	MigrationUrl string        `json:"migration_url"`
	Sum          string        `json:"sum"`
	BackupFile   string        `json:"backup_file"`
	SADevice     entity.Device `json:"-"`
}

func (req *AreaMigrationReq) ReBind(areaID uint64) (err error) {
	var (
		content []byte
		httpReq *http.Request
		jwt     string
	)
	jwt, err = GenerateMigrationJwt(MigrationClaims{
		SAID: config.GetConf().SmartAssistant.ID,
		Exp:  time.Now().Add(876000 * time.Hour).Unix(), // 设置长的时间过期时间当作永不过期
	})
	if err != nil {
		return
	}
	body := map[string]interface{}{
		"mode":                  "rebind",
		"backup_file":           req.BackupFile,
		"sum":                   req.Sum,
		"local_said":            config.GetConf().SmartAssistant.ID,
		"local_migration_token": jwt,
		"local_area_token":      setting.GetAreaAuthToken(areaID),
	}
	content, err = json.Marshal(body)
	if err != nil {
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpReq, err = http.NewRequest(http.MethodPost, req.MigrationUrl, bytes.NewBuffer(content))
	if err != nil {
		return
	}
	client := &http.Client{Timeout: HttpRequestTimeout, Transport: tr}
	// ctx, _ := context.WithTimeout(context.Background(), HttpRequestTimeout)
	// httpReq.WithContext(ctx)
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		text := fmt.Sprintf("Status Not OK, Status Code %d", httpResp.StatusCode)
		err = errors2.New(text)
		return
	}
	return
}

func (req *AreaMigrationReq) GetBackupFile() (file string, err error) {
	var (
		content    []byte
		httpReq    *http.Request
		ofile      *os.File
		fileLength int64
	)
	body := map[string]interface{}{
		"mode":        "download",
		"backup_file": req.BackupFile,
		"sum":         req.Sum,
	}
	content, err = json.Marshal(body)
	if err != nil {
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpReq, err = http.NewRequest(http.MethodPost, req.MigrationUrl, bytes.NewBuffer(content))
	if err != nil {
		return
	}
	client := &http.Client{Timeout: HttpRequestTimeout, Transport: tr}
	// ctx, _ := context.WithTimeout(context.Background(), HttpRequestTimeout)
	// httpReq.WithContext(ctx)
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		text := fmt.Sprintf("Status Not OK, Status Code %d", httpResp.StatusCode)
		err = errors2.New(text)
		return
	}
	defer httpResp.Body.Close()

	ofile, err = ioutil.TempFile(getBaseTempDir(), "temp")
	if err != nil {
		return
	}
	defer ofile.Close()
	defer func() {
		if err != nil {
			os.Remove(ofile.Name())
		}
	}()

	fileLength, err = io.Copy(ofile, httpResp.Body)
	if err != nil {
		return
	} else if fileLength != httpResp.ContentLength {
		text := fmt.Sprintf("write %d bytes, file content length %d", fileLength, httpResp.ContentLength)
		err = errors2.New(text)
		return
	}
	file = ofile.Name()

	return
}

func (req *AreaMigrationReq) ProcessCloudToLocal() (err error) {

	var (
		dir  string
		file string
	)

	file, err = req.GetBackupFile()
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	defer func() {
		os.Remove(file)
	}()

	dir, err = ioutil.TempDir(getBaseTempDir(), "temp")
	if err != nil {
		return
	}
	defer func() {
		os.RemoveAll(dir)
	}()

	err = archive.UnZip(dir, file)
	if err != nil {
		return
	}

	sa, _ := entity.GetSaDevice()
	req.SADevice = sa

	db, err := entity.OpenSqlite(path.Join(dir, "data", "smartassistant", "sadb.db"), false)
	if err != nil {
		return
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			return
		}

		sqlDB.Close()
	}()

	return entity.GetDB().Transaction(func(tx *gorm.DB) error {
		err = restoreCloudAreaDBData(db, tx)
		if err != nil {
			return err
		}

		err = req.ReBind(sa.AreaID)
		if err != nil {
			return err
		}

		return nil
	})

}

func getTableFullName(table interface{}) string {
	tableType := reflect.TypeOf(table)
	if tableType.Kind() == reflect.Ptr {
		tableType = tableType.Elem()
	}
	name := tableType.Name()
	pktPath := tableType.PkgPath()
	return fmt.Sprintf("%s.%s", pktPath, name)
}

// restoreCloudAreaDBData 从云端数据库恢复数据
func restoreCloudAreaDBData(db *gorm.DB, tx *gorm.DB) (err error) {

	saDevice := entity.Device{}
	err = tx.Model(&entity.Device{}).Where(&entity.Device{
		Model: types.SaModel,
	}).First(&saDevice).Error
	if err != nil {
		return err
	}
	err = tx.Unscoped().Delete(&entity.Area{ID: saDevice.AreaID}).Error
	if err != nil {
		return err
	}

	// 特殊处理Area表
	tx.Statement.SkipHooks = true
	defer func() {
		tx.Statement.SkipHooks = false
	}()
	area := entity.Area{}
	err = db.Model(&entity.Area{}).First(&area).Error
	if err != nil {
		return err
	}
	err = tx.Model(entity.Area{}).Create(area).Error
	if err != nil {
		return err
	}

	// 特殊处理Device表
	var devices []entity.Device
	err = db.Find(&devices).Error
	if err != nil {
		return err
	}

	if err = tx.Model(entity.Device{}).Where("true").Unscoped().Delete(nil).Error; err != nil {
		if !errors2.Is(err, gorm.ErrRecordNotFound) {
			return
		} else {
			err = nil
		}
	}
	if len(devices) > 0 {
		for index := 0; index < len(devices); index++ {
			if devices[index].Model == types.SaModel {
				continue
			}
			err = tx.Model(entity.Device{}).Create(&devices[index]).Error
			if err != nil {
				return err
			}
		}
	}
	// 更新device中sa的家庭ID
	saDevice.AreaID = area.ID
	saDevice.ID = 0
	err = tx.Model(entity.Device{}).Create(&saDevice).Error
	if err != nil {
		return err
	}

	// 遍历迁移剩余的表
	excludeTables := map[string]interface{}{
		getTableFullName(entity.Device{}): entity.Device{},
		getTableFullName(entity.Area{}):   entity.Area{},
		getTableFullName(entity.Client{}): entity.Client{},
	}
	for _, table := range entity.Tables {
		_, ok := excludeTables[getTableFullName(table)]
		if ok {
			continue
		}

		err = entity.CopyTable(db, tx, table, true)
		if err != nil {
			return
		}
	}

	return
}

func backupDatabase() {
	file := path.Join(getBaseTempDir(), fmt.Sprintf("sadb.db.%d", time.Now().Unix()))
	db, err := entity.OpenSqlite(file, false)
	if err != nil {
		logger.Debugf("open %s error %v", file, err)
		return
	}

	tx := entity.GetDB().Session(&gorm.Session{
		SkipHooks: true,
	})
	for _, table := range entity.Tables {
		err = entity.CopyTable(tx, db, table, false)
		if err != nil {
			logger.Debugf("copytable error %v", err)
			os.Remove(file)
			return
		}
	}
}

func AreaMigration(c *gin.Context) {
	var (
		req AreaMigrationReq
		err error
		saDevice entity.Device
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	backupDatabase()
	err = req.ProcessCloudToLocal()
	if err == nil {
		if saDevice, err = entity.GetSaDevice(); err != nil {
			return
		}
		err = SetAreaSynced(saDevice.AreaID)
	}

}


// SetAreaSynced 设置是否绑定云端
func SetAreaSynced(areaID uint64) (err error) {
	if err = entity.UpdateArea(areaID, map[string]interface{}{
		"is_bind_cloud": true,
	}); err != nil {
		return
	}
	return
}
