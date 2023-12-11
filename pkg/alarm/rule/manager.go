package rule

import (
	"errors"
	"github.com/sentrycloud/sentry/pkg/alarm/mysql"
	"github.com/sentrycloud/sentry/pkg/newlog"
)

type Manager struct {
	mysqlDB       *mysql.MySQL
	alarmContacts []mysql.Contact
	alarmRules    map[int]BaseRule
}

func NewManager(db *mysql.MySQL) *Manager {
	manager := new(Manager)

	manager.mysqlDB = db
	return manager
}

func (m *Manager) Start() error {
	var err error
	m.alarmContacts, err = m.mysqlDB.QueryContacts()
	if err != nil {
		return err
	}

	rules, err := m.mysqlDB.QueryAlarmRules()
	if err != nil {
		return err
	}

	m.alarmRules = make(map[int]BaseRule)
	for _, rule := range rules {
		baseRule, e := m.makeBaseRule(rule)
		if e == nil {
			m.alarmRules[rule.Id] = baseRule
			baseRule.Start()
		}
	}

	newlog.Info("total contacts=%d, total rules=%d", len(m.alarmContacts), len(m.alarmRules))
	return nil
}

func (m *Manager) makeBaseRule(rule mysql.Rule) (BaseRule, error) {
	alarmRule := AlarmRule{Rule: rule}
	var baseRule BaseRule
	switch alarmRule.RuleType {
	case AlarmTypeHeartBeat:
		baseRule = &HeartBeatRule{
			AlarmRule: alarmRule,
		}
	case AlarmTypeThreshold:
		baseRule = &ThresholdRule{
			AlarmRule: alarmRule,
		}
	case AlarmTypeTopN:
		baseRule = &TopNRule{
			AlarmRule: alarmRule,
		}
	case AlarmTypeCompare:
		baseRule = &CompareRule{
			AlarmRule: alarmRule,
		}
	default:
		newlog.Error("no such rule type: %d", alarmRule.RuleType)
		return nil, errors.New("no such rule type")
	}

	err := baseRule.Parse()
	return baseRule, err
}
