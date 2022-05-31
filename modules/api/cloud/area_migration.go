package cloud

import (
	"bytes"
	"context"
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
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/api/setting"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/archive"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/http/httpclient"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
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

func (req *AreaMigrationReq) ReBindWithContext(ctx context.Context, areaID uint64) (err error) {
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

	token, err := setting.GetAreaAuthToken(areaID)
	if err != nil {
		return
	}
	body := map[string]interface{}{
		"mode":                  "rebind",
		"backup_file":           req.BackupFile,
		"sum":                   req.Sum,
		"local_said":            config.GetConf().SmartAssistant.ID,
		"local_migration_token": jwt,
		"local_area_token":      token,
	}
	content, err = json.Marshal(body)
	if err != nil {
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, req.MigrationUrl, bytes.NewBuffer(content))
	if err != nil {
		return
	}
	client := httpclient.NewHttpClient(httpclient.WithTimeout(HttpRequestTimeout), httpclient.WithTransport(tr))
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

func (req *AreaMigrationReq) GetBackupFileWithContext(ctx context.Context) (file string, err error) {
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
	httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, req.MigrationUrl, bytes.NewBuffer(content))
	if err != nil {
		return
	}
	client := httpclient.NewHttpClient(httpclient.WithTimeout(HttpRequestTimeout), httpclient.WithTransport(tr))
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

func (req *AreaMigrationReq) ProcessCloudToLocalWithContext(ctx context.Context) (err error) {

	var (
		dir  string
		file string
	)

	file, err = req.GetBackupFileWithContext(ctx)
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

		err = req.ReBindWithContext(ctx, sa.AreaID)
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
			if devices[index].IsSa() {
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

	var clients []entity.Client
	// 查出云端的client
	if err = db.Where("area_id=?", area.ID).Find(&clients).Error; err != nil {
		return
	}

	// 云端client迁移至本地
	for _, client := range clients {
		if err = tx.Model(entity.Client{}).Create(&client).Error; err != nil {
			return
		}
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
		req      AreaMigrationReq
		err      error
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
	err = req.ProcessCloudToLocalWithContext(c.Request.Context())
	if err == nil {
		if saDevice, err = entity.GetSaDevice(); err != nil {
			return
		}
		err = SetAreaSynced(saDevice.AreaID)
	} else {

		// 迁移失败，回滚数据，需要删除添加设备后创建的家庭和SA
		// 不删除会导致无法再次添加SA

		var (
			nerr  error
			areas []entity.Area
		)
		if areas, nerr = entity.GetAreas(); nerr != nil {
			logger.Warnf("area migration rollback, get areas error %v", nerr)
			return
		}

		// 删除家庭，删除家庭会将家庭的所有设备删除

		for _, area := range areas {
			if nerr = entity.DelAreaByID(area.ID); nerr != nil {
				logger.Warnf("area migration rollback, del area %d error %v", area.ID, nerr)
			}
		}
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
