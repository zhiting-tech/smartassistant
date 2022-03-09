package area

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/clouddisk"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/modules/utils/url"
	plugin2 "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
	"strings"

	"github.com/zhiting-tech/smartassistant/modules/api/setting"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/archive"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
)

type AreaDelSaReq struct {
	AreaID  uint64	`uri:"id" binding:"required"`
	IsMigrationSA   bool `json:"is_migration_sa"`
	IsDelCloudDisk	bool `json:"is_del_cloud_disk"`
	CloudAreaId		string `json:"cloud_area_id"`
	CloudAccessToken string `json:"cloud_access_token"`
}

func (req *AreaDelSaReq) validate(userID int)(err error) {
	if _, err = entity.GetAreaByID(req.AreaID); err != nil {
		return
	}
	if !entity.IsOwnerOfArea(userID, req.AreaID) {
		err = errors.New(status.Deny)
		return
	}
	if !req.IsMigrationSA {
		return nil
	}

	if req.CloudAccessToken == "" || req.CloudAreaId == "" {
		err = errors.New(errors.BadRequest)
		return
	}
	return
}

func AreaDelSa(c *gin.Context) {
	var (
		err error
		req AreaDelSaReq
		resp clouddisk.DelAreaStatus
		fileName string
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()
	if err = c.BindUri(&req); err != nil {
		return
	}
	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if err = req.validate(session.Get(c).UserID); err != nil {
		return
	}

	// 向云端家庭迁移本地sa数据
	if req.IsMigrationSA {
		// 先设置智汀设备服务为云端
		if err = setDeviceServer(req.AreaID, req.CloudAreaId, req.CloudAccessToken); err != nil {
			return
		}
		fileName, err = backupDatabase()
		if err != nil {
			return
		}
		defer os.Remove(fileName)
		if err = dbFileUploadToSC(fileName, req.CloudAreaId, req.CloudAccessToken); err != nil {
			return
		}
	}
	// 删除网盘
	resp, err = ProcessDelArea(c, req.AreaID, req.IsDelCloudDisk)
	return
}

// backupDatabase 备份本地数据库
func backupDatabase() (zipFileName string, err error){
	fileName := path.Join(config.GetConf().SmartAssistant.BackupPath(), "sadb.db")
	db, err := entity.OpenSqlite(fileName, false)
	if err != nil {
		logger.Debugf("open %s error %v", fileName, err)
		return
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			return
		}
		sqlDB.Close()
		os.Remove(fileName)
	}()

	tx := entity.GetDB()
	db.Statement.SkipHooks = true
	for _, table := range entity.Tables {
		if _, ok := table.(entity.TaskLog); ok {
			continue
		}
		err = entity.CopyTable(tx, db, table, false)
		if err != nil {
			return
		}
	}
	zipFileName = fmt.Sprintf("%s.zip", path.Join(config.GetConf().SmartAssistant.BackupPath(), "sadb"))
	if err = archive.Zip(zipFileName, fileName); err != nil {
		os.Remove(zipFileName)
		return
	}
	return
}

// dbFileUploadToSC 上传备份的数据文件到sc
func dbFileUploadToSC(fileName, cloudAreaID, cloudAccessToken string) (err error){
	saID := config.GetConf().SmartAssistant.ID
	scUrl := config.GetConf().SmartCloud.URL()
	reqUrl := fmt.Sprintf("%s/sa/%s/migration_cloud/%s", scUrl, saID, cloudAreaID)
	req, err := genMigrationCloudRequest(fileName, reqUrl, cloudAccessToken)
	if err != nil {
		return
	}
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Warnf("request %s error %v\n", reqUrl, err)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		logger.Warnf("request %s error,status:%v\n", reqUrl, httpResp.Status)
		return
	}
	defer httpResp.Body.Close()
	respBytes, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return
	}
	var data response.BaseResponse
	if err = json.Unmarshal(respBytes, &data); err != nil {
		return
	}
	if data.Status != 0 {
		logger.Warnf("response %s error, status is %d, reason is %s", reqUrl, data.Status, data.Reason)
		err = errors.New(errors.InternalServerErr)
		return
	}
	return
}

// genMigrationCloudRequest 生成迁移请求
func genMigrationCloudRequest(fileName, reqUrl, cloudAccessToken string) (req *http.Request, err error) {
	// 写入文件数据，token字段
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		return
	}
	if err = bodyWriter.WriteField("cloud_access_token", cloudAccessToken); err != nil {

		return
	}
	fileWrite, err := bodyWriter.CreateFormFile("file_upload", file.Name())
	if err != nil {
		return
	}
	_, err = io.Copy(fileWrite, file)
	if err != nil {
		return
	}
	// 需要关闭, 数据才会写入
	bodyWriter.Close()
	logger.Debug(reqUrl)
	req, err = http.NewRequest(http.MethodPost, reqUrl, bodyBuf)
	if err != nil {
		logger.Warnf("NewRequest error %v\n", err)
		return
	}
	req.Header = cloud.GetCloudReqHeader()
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	ctx, _ := context.WithTimeout(context.Background(), setting.HttpRequestTimeout)
	req.WithContext(ctx)
	return
}

// setDeviceServer 设置美汇智居设备服务
func setDeviceServer(currentAreaID uint64, cloudAreaID, cloudAccessToken string)(err error) {
	var devices []entity.Device
	if devices, err = entity.GetDevices(currentAreaID); err != nil {
		return
	}

	addrInfo := strings.Split(config.GetConf().SmartCloud.Domain, ":")
	if len(addrInfo) == 0 {
		err = errors.New(errors.InternalServerErr)
		return
	}

	query := map[string]interface{}{
		"server": fmt.Sprintf("%s:%d", addrInfo[0], 54321),
		"area_id": cloudAreaID,
		"access_token": cloudAccessToken,
	}
	serverInfo := url.Join(url.BuildQuery(query))

	for _, d := range devices {
		if d.Model == types.SaModel || !entity.IsMeiHuiZhiJuBrand(d.PluginID){
			continue
		}
		if !plugin.GetGlobalClient().IsOnline(d) {
			continue
		}
		var das plugin.DeviceInstances
		das, err = device.GetThingModel(d)
		if err != nil {
			logger.Warnf(" device getThingModel error is %s", err.Error())
			continue
		}
		for _, instance := range das.Instances {
			if instance.Type != "info" {
				continue
			}
			req := plugin2.SetRequest{
				Attributes: []plugin2.SetAttribute{{
					InstanceID: instance.InstanceId,
					Attribute: "server_info",
					Val: serverInfo,
				}},
			}
			data, _ := json.Marshal(req)
			_, err = plugin.GetGlobalClient().SetAttributes(d, data)
			if err != nil {
				logger.Warnf("set attributes error is %s", err.Error())
			}
			break
		}
	}
	return
}