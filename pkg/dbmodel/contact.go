package dbmodel

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
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

func QueryAllContacts() ([]AlarmContact, error) {
	var contacts []AlarmContact
	result := db.Where("is_deleted = ?", 0).Find(&contacts)
	if result.Error != nil {
		newlog.Error("query alarm contacts failed: %v", result.Error)
		return nil, result.Error
	}

	return contacts, nil
}

func AddContact(contact *AlarmContact) error {
	contact.SetTimeNow()
	result := db.Select("name", "phone", "mail", "wechat").Create(contact)
	return result.Error
}

func UpdateContact(contact *AlarmContact) error {
	result := db.Model(contact).Select("name", "phone", "mail", "wechat").Updates(contact)
	return result.Error
}

func DeleteContact(contact *AlarmContact) error {
	// soft delete
	result := db.Model(contact).Update("is_deleted", 1)
	return result.Error
}
