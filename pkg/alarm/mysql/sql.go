package mysql

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/alarm/config"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

type MySQL struct {
	handle *gorm.DB
}

const (
	AlarmContactSQL = "select `name`, `phone`, `mail`, `wechat` from `alarm_contact` where `deleted`=0;"
	AlarmRuleSQL    = "select `id`, `name`, `type`, `query_range`, `contacts`, `level`, `message`, `data_source`, `trigger` from `alarm_rule` where `deleted`=0;"
)

type Contact struct {
	ID      uint
	Name    string
	Phone   string
	Mail    string
	Wechat  string
	Deleted int
	Created time.Time
	Updated time.Time
}

type Rule struct {
	Id         int
	Name       string
	RuleType   int `gorm:"column:type"`
	QueryRange int // query interval is half of query range
	Contacts   string
	Level      int
	Message    string
	DataSource string
	Trigger    string
	Deleted    int
	Created    time.Time
	Updated    time.Time
}

func NewMySQL(c *config.MySQLConfig) (*MySQL, error) {
	db := new(MySQL)

	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", c.Username, c.Password, c.Host, c.Port, c.DBName)
	db.handle, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		newlog.Error("open db connections failed: %v", err)
		return nil, err
	}

	sqlDB, _ := db.handle.DB()
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	return db, nil
}

func (db *MySQL) QueryContacts() ([]Contact, error) {
	var contacts []Contact
	result := db.handle.Table("alarm_contact").Where("deleted = ?", 0).Find(&contacts)
	if result.Error != nil {
		newlog.Error("query alarm contacts failed: %v", result.Error)
		return nil, result.Error
	}

	return contacts, nil
}

func (db *MySQL) QueryAlarmRules() ([]Rule, error) {
	var rules []Rule
	result := db.handle.Table("alarm_rule").Where("deleted = ?", 0).Find(&rules)
	if result.Error != nil {
		newlog.Error("query alarm contacts failed: %v", result.Error)
		return nil, result.Error
	}

	return rules, nil
}
