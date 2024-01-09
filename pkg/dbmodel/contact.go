package dbmodel

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"time"
)

type AlarmContact struct {
	Entity
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Mail   string `json:"mail"`
	Wechat string `json:"wechat"`
}

func (AlarmContact) TableName() string {
	return "alarm_contact"
}

func QueryUpdateContacts(updated time.Time, contacts *[]AlarmContact) error {
	result := db.Where("updated >= ?", updated).Find(contacts)
	return result.Error
}

func IsContactNameExist(name string) bool {
	var contacts []AlarmContact
	result := db.Where("is_deleted=? and name=?", 0, name).Find(&contacts)
	if result.Error != nil {
		newlog.Error("query name in contact failed: %v", result.Error)
		return true
	}

	return len(contacts) > 0
}
