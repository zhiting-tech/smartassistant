package setting

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type GetSettingQuery struct {
	UserCredentialFoundSetting    *entity.UserCredentialFoundSetting    `json:"user_credential_found_setting"`
	LogSetting                    *entity.LogSwitchSetting              `json:"log_setting"`
	GetCloudDiskCredentialSetting *entity.GetCloudDiskCredentialSetting `json:"get_cloud_disk_credential_setting"`
}

// UpdateSetting 修改全局配置
func UpdateSetting(c *gin.Context) {

	var (
		req         GetSettingQuery
		err         error
		sessionUser *session.User
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	// 只有SA拥有者才能设置
	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	// 修改是否允许找回用户凭证的配置
	if req.UserCredentialFoundSetting != nil {
		err = req.UpdateUserCredentialFound(sessionUser.AreaID)
		if err != nil {
			return
		}
	}

	// 修改是否打开日志上传
	if req.LogSetting != nil {
		err = req.UpdateLogSetting(sessionUser.AreaID)
		if err != nil {
			return
		}
	}

	// 修改是否允许获取网盘凭证的配置
	if req.GetCloudDiskCredentialSetting != nil {
		if err = req.UpdateCloudDiskCredential(sessionUser.AreaID); err != nil {
			return
		}
	}
}

func (req *GetSettingQuery) UpdateUserCredentialFound(areaID uint64) (err error) {

	// 更新是否允许找回用户凭证
	setting := req.UserCredentialFoundSetting
	err = entity.UpdateSetting(entity.UserCredentialFoundType, &setting, areaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	// 发送SC请求SA的认证token给SC
	if setting.UserCredentialFound {
		go sendAreaAuthToSC(areaID)
	}

	return nil
}

type LogSetting struct {
	TLS       bool
	LogSwitch bool
	Domain    string
	ID        string
	Key       string
}

var fluentdConfigTemplate = `
<source>
  @type forward
</source>

<filter **>
  @type parser
  key_name log
  reserve_time true
  emit_invalid_record_to_error false
  <parse>
    @type json
    time_key time
    time_type string
    time_format %Y-%m-%dT%H:%M:%S
    keep_time_key true
  </parse>
</filter>

<filter **>
  @type record_transformer
  <record>
    sa_id {{ .ID }}
    tag ${record["app"]}.${record["module"]}
  </record>
</filter>

<match **>
  @type copy
  <store>
    @type file
    <format>
      @type json
    </format>
    path /var/log/smartassistant
    flush_interval 1s
    append true
  </store>
{{if .LogSwitch}}
  <store>
    @type http
    endpoint  {{if .TLS}}https{{else}}http{{end}}://{{ .Domain }}/api/log_replay
    open_timeout 2
    http_method post

    <format>
      @type json
    </format>
    <buffer>
      flush_interval 5s
    </buffer>
    <auth>
      method basic
      username {{ .ID }}
      password {{ .Key }}
    </auth>
  </store>
{{end}}
</match>
`

func (req *GetSettingQuery) UpdateLogSetting(areaID uint64) (err error) {
	setting := req.LogSetting
	err = entity.UpdateSetting(entity.LogSwitch, &setting, areaID)

	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	conf := config.GetConf()
	ls := LogSetting{
		TLS:       conf.SmartCloud.TLS,
		ID:        conf.SmartAssistant.ID,
		Key:       conf.SmartAssistant.Key,
		Domain:    conf.SmartCloud.Domain,
		LogSwitch: setting.LogSwitch,
	}
	// 获取根目录路径
	dir := config.GetConf().SmartAssistant.RuntimePath
	// 拼接文件路径
	f, err := os.Create(filepath.Join(dir, "config", "fluentd.conf"))
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	defer f.Close()
	tmpl, err := template.New("log").Parse(fluentdConfigTemplate)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	err = tmpl.Execute(f, ls)
	if err != nil {
		panic(err)
	}

	// 获取配置文件的log镜像名
	dockerList, _ := docker.GetClient().ContainerList()
	for _, v := range dockerList {
		// 查找镜像名是否包含fluentd
		if strings.Contains(v.Image, "fluentd") {
			// 发送重新加载配置文件信号
			err = docker.GetClient().ContainerKillByImage(v.Image, "SIGUSR2")
			// 发送失败
			if err != nil {
				err = errors.Wrap(err, errors.InternalServerErr)
				return
			}
		}
	}

	return nil
}

func (req *GetSettingQuery) UpdateCloudDiskCredential(areaID uint64) (err error) {
	// 更新是否允许获取网盘凭证
	setting := req.GetCloudDiskCredentialSetting
	err = entity.UpdateSetting(entity.GetCloudDiskCredential, &setting, areaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 发送sc请求sa的认证token给SC
	if setting.GetCloudDiskCredentialSetting {
		go sendAreaAuthToSC(areaID)
	}
	return
}
