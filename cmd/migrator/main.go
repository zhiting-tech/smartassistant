package main

import (
	"encoding/json"
	"flag"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
	"gopkg.in/oauth2.v3"
	"gorm.io/gorm"
)

var configFile = flag.String("c", "/mnt/data/zt-smartassistant/config/smartassistant.yaml", "config file")

func main() {
	flag.Parse()
	_ = config.InitConfig(*configFile)

	if err := entity.GetDB().Transaction(func(tx *gorm.DB) error {
		// 更新clients
		if err := migratorClient(tx); err != nil {
			return err
		}
		// 更新sa
		if err := updateSAModel(tx); err != nil {
			return err
		}

		// 更新除SA之外的设备
		if err := updateDeviceWithOutSA(tx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logger.Error(" err: ", err)
		logger.Println("migrator fail....")
		return
	}
	logger.Println("migrator success....")
}

func updateSAModel(db *gorm.DB) error {
	if err := db.Model(&entity.Device{}).Where("model=?", "smart_assistant").UpdateColumn("model", types.SaModel).Error; err != nil {
		logger.Error("update sa err: ", err)
		return err
	}

	return nil
}

type Device struct {
	entity.Device
	Identity string
}

var tm = &thingmodel.ThingModel{
	Instances: []thingmodel.Instance{
		thingmodel.Instance{
			IID: "7",
			Services: []thingmodel.Service{
				thingmodel.Service{
					Type: "info",
					Attributes: []thingmodel.Attribute{
						{
							AID:        1,
							Type:       "manufacturer",
							Permission: 1,
							ValType:    "string",
							Val:        "zhiting",
						},
						{
							AID:        2,
							Type:       "identify",
							Permission: 1,
							ValType:    "string",
							Val:        "84f703a6b8b5",
						},
						{
							AID:        3,
							Type:       "model",
							Permission: 1,
							ValType:    "string",
							Val:        "MH-SW3ZLW001W",
						},
						{
							AID:        4,
							Type:       "version",
							Permission: 1,
							ValType:    "string",
							Val:        "version",
						},
						{
							AID:        5,
							Type:       "server_info",
							Permission: 1,
							ValType:    "string",
							Val:        "",
						},
					},
				},
				thingmodel.Service{
					Type: "switch",
					Attributes: []thingmodel.Attribute{
						thingmodel.Attribute{
							AID:        6,
							Type:       "on_off",
							Permission: 7,
							ValType:    "string",
							Val:        "on",
						},
					},
				},
				thingmodel.Service{
					Type: "switch",
					Attributes: []thingmodel.Attribute{
						thingmodel.Attribute{
							AID:        7,
							Type:       "on_off",
							Permission: 7,
							ValType:    "string",
							Val:        "on",
						},
					},
				},
				thingmodel.Service{
					Type: "switch",
					Attributes: []thingmodel.Attribute{
						thingmodel.Attribute{
							AID:        8,
							Type:       "on_off",
							Permission: 7,
							ValType:    "string",
							Val:        "on",
						},
					},
				},
			},
		},
	},
	OTASupport: false,
}

func updateDeviceWithOutSA(db *gorm.DB) error {
	var devices []Device

	if err := db.Select("id", "identity", "model").Where("model != ?", types.SaModel).Find(&devices).Error; err != nil {
		return err
	}

	for _, d := range devices {
		if err := db.Model(&entity.Device{ID: d.ID}).UpdateColumn("iid", d.Identity).Error; err != nil {
			return err
		}

		for _, i := range tm.Instances {
			tm.Instances[0].IID = d.Identity
			for j, s := range i.Services {
				if s.Type == "info" {
					for k, a := range s.Attributes {
						if a.Type == "identify" {
							tm.Instances[0].Services[j].Attributes[k].Val = d.Identity
						}
					}
				}
			}
		}

		dJson, err := json.Marshal(tm)
		if err != nil {
			return err
		}

		if d.Model == "MH-SW3ZLW001W" {

			if err := db.Model(&entity.Device{ID: d.ID}).UpdateColumn("thing_model", dJson).Error; err != nil {
				logger.Error("update thing model err: ", err)
				return err
			}

		}

	}
	return nil
}

func migratorClient(tx *gorm.DB) error {

	// 查出数据库中已有的client
	var clients []entity.Client
	if err := tx.Find(&clients).Order("id asc").Error; err != nil {
		logger.Errorf("find clients err: %v", err)
		return err
	}

	// 组装成saClient, scClient
	var saClient, scClient entity.Client
	for _, client := range clients {
		if !(client.GrantType == string(oauth2.ClientCredentials)) {
			saClient = entity.Client{
				ClientID:     client.ClientID,
				ClientSecret: client.ClientSecret,
				GrantType:    client.GrantType,
				AllowScope:   client.AllowScope,
			}
			continue
		}

		scClient = entity.Client{
			ClientID:     client.ClientID,
			ClientSecret: client.ClientSecret,
			GrantType:    client.GrantType,
			AllowScope:   client.AllowScope,
		}
	}

	// 查出数据库中现有的家庭
	var areas []entity.Area
	if err := tx.Find(&areas).Error; err != nil {
		logger.Errorf("find areas err: %v", err)
		return err
	}

	// 为对应家庭包装saClient和scClient
	var saClientsOfArea []entity.Client
	var scClientsOfArea []entity.Client
	for _, area := range areas {
		saClientOfArea := entity.Client{
			AreaID:       area.ID,
			ClientID:     saClient.ClientID,
			ClientSecret: saClient.ClientSecret,
			GrantType:    saClient.GrantType,
			AllowScope:   saClient.AllowScope,
			Type:         entity.AreaClient,
		}
		saClientsOfArea = append(saClientsOfArea, saClientOfArea)

		scClientOfArea := entity.Client{
			AreaID:       area.ID,
			ClientID:     scClient.ClientID,
			ClientSecret: scClient.ClientSecret,
			GrantType:    scClient.GrantType,
			AllowScope:   scClient.AllowScope,
			Type:         entity.SCClient,
		}

		scClientsOfArea = append(scClientsOfArea, scClientOfArea)
	}

	var newClients []entity.Client
	newClients = append(newClients, saClientsOfArea...)
	newClients = append(newClients, scClientsOfArea...)

	// 移除唯一索引
	tx.Exec("drop index idx_clients_client_secret;")
	tx.Exec("drop index idx_clients_client_id;")

	// 批量插入client
	if err := tx.Session(&gorm.Session{SkipHooks: true}).CreateInBatches(&newClients, len(newClients)).Error; err != nil {
		logger.Errorf("create clients err: %v", err)
		return err
	}

	// 删除原有的client
	if err := tx.Model(&entity.Client{}).Delete(clients).Error; err != nil {
		logger.Error("delete clients err: %v", err)
		return err
	}
	return nil
}
