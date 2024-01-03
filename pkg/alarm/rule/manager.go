package rule

import (
	"errors"
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
)

type Manager struct {
	alarmContacts []dbmodel.AlarmContact
	alarmRules    map[uint32]BaseRule
}

func NewManager() *Manager {
	manager := new(Manager)
	return manager
}

func (m *Manager) Start() error {
	var err error
	err = dbmodel.QueryAllEntity(&m.alarmContacts)
	if err != nil {
		return err
	}

	var rules []dbmodel.AlarmRule
	err = dbmodel.QueryAllEntity(&rules)
	if err != nil {
		return err
	}

	m.alarmRules = make(map[uint32]BaseRule)
	for _, rule := range rules {
		baseRule, e := m.makeBaseRule(rule)
		if e == nil {
			m.alarmRules[rule.ID] = baseRule
			baseRule.Start()
		}
	}

	newlog.Info("total contacts=%d, total rules=%d", len(m.alarmContacts), len(m.alarmRules))
	return nil
}

func (m *Manager) makeBaseRule(rule dbmodel.AlarmRule) (BaseRule, error) {
	alarmRule := AlarmRule{AlarmRule: rule}
	var baseRule BaseRule
	switch alarmRule.Type {
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
		newlog.Error("no such rule type: %d", alarmRule.Type)
		return nil, errors.New("no such rule type")
	}

	err := baseRule.Parse()
	return baseRule, err
}
