package dbmodel

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

// constrain all gorm usage in this db_model package
// use gorm conventions: https://gorm.io/docs/conventions.html
var db *gorm.DB

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

func NewMySQL(c *MySQLConfig) error {
	if db != nil {
		return nil
	}

	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", c.Username, c.Password, c.Host, c.Port, c.DBName)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		newlog.Error("open db connections failed: %v", err)
		return err
	}

	sqlDB, err := db.DB()

	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	return nil
}
