package entity

import (
	"github.com/twinj/uuid"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/rand"
	"gopkg.in/oauth2.v3"
	"gorm.io/gorm"
	"strings"
)

type Client struct {
	ID           int
	Name         string `gorm:"UniqueIndex"`
	ClientID     string `gorm:"UniqueIndex"`
	ClientSecret string `gorm:"UniqueIndex"`
	GrantType    string
	AllowScope   string // 允许客户端申请的权限
}

func (c Client) TableName() string {
	return "clients"
}

func GetDefaultScope() (scope string) {
	scopeList := make([]string, 0)
	for k := range types.Scopes {
		scopeList = append(scopeList, k)
	}
	return strings.Join(scopeList, ",")
}

// CreateClient 创建应用
func CreateClient(grantType, allowScope, name string) (client Client, err error) {
	client = Client{
		GrantType:  getAllowGrantType(grantType),
		AllowScope: allowScope,
		Name:       name,
	}

	if err = GetDB().FirstOrCreate(&client).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

// GetClientByClientID 根据ClientID获取Client信息
func GetClientByClientID(clientID string) (client Client, err error) {
	if err = GetDB().Where("client_id=?", clientID).First(&client).Error; err != nil {
		return
	}
	return
}

// getAllowGrantType 获取Client允许的授权类型
func getAllowGrantType(grantType string) string {
	if grantType != string(oauth2.Implicit) || grantType != string(oauth2.ClientCredentials) {
		grantType = grantType + "," + string(oauth2.Refreshing)
	}
	return grantType
}

// InitClient 初始化Client
func InitClient() (err error) {
	var clients = make([]Client, 0)
	saClient := Client{
		Name:       config.GetConf().SmartAssistant.ID,
		GrantType:  string(oauth2.Implicit) + "," + string(oauth2.PasswordCredentials) + "," + string(oauth2.Refreshing),
		AllowScope: GetDefaultScope(),
	}

	scClient := Client{
		Name:       config.GetConf().SmartCloud.Domain,
		GrantType:  string(oauth2.ClientCredentials),
		AllowScope: GetDefaultScope(),
	}

	clients = append(clients, saClient, scClient)

	for _, client := range clients {
		if err = GetDB().FirstOrCreate(&client, Client{Name: client.Name}).Error; err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return
		}
	}
	return
}

// GetSAClient 获取SAClient
func GetSAClient() (client Client, err error) {
	if err = GetDB().First(&client, "name=?", config.GetConf().SmartAssistant.ID).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

// GetSCClient 获取SCClient
func GetSCClient() (client Client, err error) {
	if err = GetDB().First(&client, "name=?", config.GetConf().SmartCloud.Domain).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

func (c *Client) BeforeCreate(tx *gorm.DB) error {
	c.ClientID = uuid.NewV4().String()
	c.ClientSecret = rand.StringK(32, rand.KindAll)
	return nil
}
