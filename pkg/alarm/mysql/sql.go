package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sentrycloud/sentry/pkg/alarm/config"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"time"
)

type MySQL struct {
	handle *sql.DB
}

const (
	AlarmContactSQL = "select `name`, `phone`, `mail`, `wechat` from `alarm_contact` where `deleted`=0;"
	AlarmRuleSQL    = "select `id`, `name`, `type`, `query_range`, `contacts`, `level`, `message`, `data_source`, `trigger` from `alarm_rule` where `deleted`=0;"
)

type Contact struct {
	Name   string
	Phone  string
	Mail   string
	WeChat string
}

type Rule struct {
	Id         int
	Name       string
	RuleType   int
	QueryRange int // query interval is half of query range
	Contacts   string
	Level      int
	Message    string
	DataSource string
	Trigger    string
}

func NewMySQL(c *config.MySQLConfig) (*MySQL, error) {
	db := new(MySQL)

	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.Username, c.Password, c.Host, c.Port, c.DBName)
	db.handle, err = sql.Open("mysql", dsn)
	if err != nil {
		newlog.Error("open db connections failed: %v", err)
		return nil, err
	}

	db.handle.SetConnMaxLifetime(1 * time.Minute)
	db.handle.SetMaxOpenConns(10)
	db.handle.SetMaxIdleConns(10)
	return db, nil
}

func (db *MySQL) QueryContacts() (map[string]Contact, error) {
	rows, err := db.handle.Query(AlarmContactSQL)
	if err != nil {
		newlog.Error("query alarm contacts failed: %v", err)
		return nil, err
	}

	defer rows.Close()
	alarmContacts := make(map[string]Contact)
	for rows.Next() {
		var contact Contact
		err = rows.Scan(&contact.Name, &contact.Phone, &contact.Mail, &contact.WeChat)
		if err != nil {
			newlog.Error("scan alarm contact failed: %v", err)
			return nil, err
		}

		alarmContacts[contact.Name] = contact
	}

	return alarmContacts, nil
}

func (db *MySQL) QueryAlarmRules() (map[int]Rule, error) {
	rows, err := db.handle.Query(AlarmRuleSQL)
	if err != nil {
		newlog.Error("query alarm contacts failed: %v", err)
		return nil, err
	}

	defer rows.Close()
	alarmRules := make(map[int]Rule)
	for rows.Next() {
		var rule Rule
		err = rows.Scan(&rule.Id, &rule.Name, &rule.RuleType, &rule.QueryRange, &rule.Contacts, &rule.Level, &rule.Message, &rule.DataSource, &rule.Trigger)
		if err != nil {
			newlog.Error("scan alarm contact failed: %v", err)
			return nil, err
		}

		alarmRules[rule.Id] = rule
	}

	return alarmRules, nil
}
