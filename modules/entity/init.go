package entity

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	logger1 "github.com/zhiting-tech/smartassistant/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zhiting-tech/smartassistant/modules/config"
)

var db *gorm.DB
var once sync.Once

var Tables []interface{} = []interface{}{
	Device{}, Location{}, Area{}, Role{}, RolePermission{},
	User{}, UserRole{}, Scene{}, SceneCondition{},
	SceneTask{}, TaskLog{}, GlobalSetting{}, PluginInfo{}, Client{},
	Department{}, DepartmentUser{}, DeviceState{}, FileInfo{}, BackupInfo{},
	UserCommonDevice{},
}

func GetDB() *gorm.DB {
	once.Do(func() {
		loadDB()
	})
	return db
}

func loadDB() {

	var dialect gorm.Dialector
	database := config.GetConf().SmartAssistant.Database
	driver := database.Driver
	var dsn string
	switch driver {
	case "sqlite":
		dsn = filepath.Join(config.GetConf().SmartAssistant.DataPath(),
			"smartassistant", "sadb.db")
		dialect = Open(dsn)
	case "postgres", "postgresql":
		format := "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"
		dsn = fmt.Sprintf(format, database.Host, database.Port, database.Username,
			database.Password, database.Name, "disable")
		dialect = postgres.Open(dsn)
	default:
		panic(fmt.Errorf("invalid dialector %v", driver))
	}
	sqldb, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("数据库连接失败 %v，dsn: %s", err.Error(), dsn))
	}
	// PRAGMA foreign_keys=ON 开启外键关联约束
	sqldb.Exec("PRAGMA foreign_keys=ON;")

	logMode := logger.Warn
	if config.GetConf().Debug {
		logMode = logger.Info
	}
	db = sqldb.Session(&gorm.Session{
		Logger: logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logMode,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
	})

	if err = db.AutoMigrate(Tables...); err != nil {
		logger1.Panicf("migrate err:%s", err.Error())
	}
}

func OpenSqlite(path string, enableForeign bool) (*gorm.DB, error) {
	dialect := Open(path)
	sqldb, err := gorm.Open(dialect, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: !enableForeign,
	})
	if err != nil {
		return nil, err
	}
	if enableForeign {
		sqldb.Exec("PRAGMA foreign_keys=ON;")
	}

	logMode := logger.Warn
	if config.GetConf().Debug {
		logMode = logger.Info
	}
	sess := sqldb.Session(&gorm.Session{
		Logger: logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logMode,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
	})

	if err = sess.AutoMigrate(Tables...); err != nil {
		return nil, err
	}

	return sess, nil
}

// FromArea 查找家庭对应的数据
func FromArea(areaID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("area_id=?", areaID)
	}
}

func GetDBWithAreaScope(areaID uint64) *gorm.DB {
	return GetDB().Scopes(FromArea(areaID))
}

func GetDBWithAreaScopeTx(tx *gorm.DB, areaID uint64) *gorm.DB {
	return tx.Scopes(FromArea(areaID))
}
