package contact

import (
	"github.com/sentrycloud/sentry/pkg/alarm/schedule"
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"regexp"
	"sync"
	"time"
)

const (
	updateContactInterval    = 30 * time.Second
	maxDelayForUpdateContact = 6 * time.Minute
)

const (
	PhoneMsg = iota
	WechatMsg
	MailMsg
)

var (
	// contact id will not change, but name may be changed, so we need 2 map to store the contact information
	// one for update and another for query
	contactsIdMap    sync.Map
	contactsNameMap  sync.Map
	latestUpdateTime time.Time
	separatorRegexp  *regexp.Regexp
)

func Init() error {
	var alarmContacts []dbmodel.AlarmContact
	err := dbmodel.QueryAllEntity(&alarmContacts)
	if err != nil {
		return err
	}

	for _, contact := range alarmContacts {
		updateLatestTime(contact.Updated)
		contactsIdMap.Store(contact.ID, &contact)
		contactsNameMap.Store(contact.Name, &contact)
	}

	separatorRegexp = regexp.MustCompile("[,; ]")
	schedule.Repeat(updateContactInterval, updateContacts)
	return nil
}

func GetAlarmContact(contacts string, senderType int) []string {
	contactNames := separatorRegexp.Split(contacts, -1)

	var senderContacts []string
	for _, name := range contactNames {
		contactPtr, exist := contactsNameMap.Load(name)
		if exist && contactPtr != nil {
			contact := contactPtr.(*dbmodel.AlarmContact)
			if senderType == PhoneMsg {
				senderContacts = append(senderContacts, contact.Phone)
			} else if senderType == WechatMsg {
				senderContacts = append(senderContacts, contact.Wechat)
			} else {
				senderContacts = append(senderContacts, contact.Mail)
			}
		}
	}

	return senderContacts
}

func updateLatestTime(t time.Time) {
	if latestUpdateTime.Before(t) {
		latestUpdateTime = t
	}
}

func updateContacts() {
	now := time.Now()
	if now.Sub(latestUpdateTime) > maxDelayForUpdateContact {
		latestUpdateTime = now.Add(-maxDelayForUpdateContact)
	}

	var contacts []dbmodel.AlarmContact
	err := dbmodel.QueryUpdateContacts(latestUpdateTime, &contacts)
	if err != nil {
		newlog.Error("QueryUpdateContacts failed: %v", err)
		return
	}

	for _, contact := range contacts {
		updateLatestTime(contact.Updated)

		existContactPtr, exist := contactsIdMap.Load(contact.ID)
		if !exist {
			if contact.IsDeleted == 0 {
				newlog.Info("add contact, id=%d, name=%s", contact.ID, contact.Name)
				contactsIdMap.Store(contact.ID, &contact)
				contactsNameMap.Store(contact.Name, &contact)
			}
		} else {
			if contact.IsDeleted == 0 {
				existContact := existContactPtr.(*dbmodel.AlarmContact)
				if contact.Updated.Compare(existContact.Updated) > 0 {
					newlog.Info("update contact, id=%d, name=%s", contact.ID, contact.Name)
					contactsIdMap.Store(contact.ID, &contact)
					contactsNameMap.Store(contact.Name, &contact)
				}
			} else {
				newlog.Info("delete contact, id=%d, name=%s", contact.ID, contact.Name)
				contactsIdMap.Delete(contact.ID)
				contactsNameMap.Delete(contact.Name)
			}
		}
	}
}
