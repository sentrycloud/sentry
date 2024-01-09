package rule

import (
	"errors"
	"github.com/RussellLuo/timingwheel"
	"github.com/sentrycloud/sentry/pkg/alarm/schedule"
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"time"
)

const (
	UpdateRuleInterval        = 30 * time.Second
	MaxDelayForUpdateRuleTime = 10 * time.Minute
)

type Manager struct {
	alarmRules       map[uint32]BaseRule // only schedule and init goroutine will change this, so it's safe to not use sync.Map
	latestUpdateTime time.Time
	updateRuleTimer  *timingwheel.Timer
}

func NewManager() *Manager {
	manager := new(Manager)
	return manager
}

func (m *Manager) Start() error {
	var rules []dbmodel.AlarmRule
	err := dbmodel.QueryAllEntity(&rules)
	if err != nil {
		return err
	}

	m.alarmRules = make(map[uint32]BaseRule)
	for _, rule := range rules {
		m.updateLatestTime(rule.Updated)
		m.addNewRule(rule)
	}

	m.updateRuleTimer = schedule.Repeat(UpdateRuleInterval, m.updateRules)
	newlog.Info("start total rules=%d", len(m.alarmRules))
	return nil
}

func (m *Manager) updateLatestTime(ruleUpdateTime time.Time) {
	if m.latestUpdateTime.Before(ruleUpdateTime) {
		m.latestUpdateTime = ruleUpdateTime
	}
}

func (m *Manager) updateRules() {
	var rules []dbmodel.AlarmRule

	now := time.Now()
	if now.Sub(m.latestUpdateTime) > MaxDelayForUpdateRuleTime {
		m.latestUpdateTime = now.Add(-MaxDelayForUpdateRuleTime)
	}

	err := dbmodel.QueryUpdateRules(m.latestUpdateTime, &rules)
	if err != nil {
		newlog.Error("query updated rule failed: %v", err)
		return
	}

	for _, rule := range rules {
		m.updateLatestTime(rule.Updated)

		existRule, exist := m.alarmRules[rule.ID]
		if !exist {
			if rule.IsDeleted == 0 {
				m.addNewRule(rule)
			}
		} else {
			if rule.IsDeleted == 0 {
				if rule.Updated.Compare(existRule.GetAlarmRule().Updated) > 0 {
					newlog.Info("update exist rule, ruleId=%d, name=%s, type=%d, range=%d", rule.ID, rule.Name, rule.Type, rule.QueryRange)
					existRule.Stop()
					delete(m.alarmRules, rule.ID)
					m.addNewRule(rule)
				}
			} else {
				newlog.Info("delete exist rule, ruleId=%d, name=%s, type=%d, range=%d", rule.ID, rule.Name, rule.Type, rule.QueryRange)
				existRule.Stop()
				delete(m.alarmRules, rule.ID)
			}
		}
	}
}

func (m *Manager) addNewRule(rule dbmodel.AlarmRule) {
	baseRule, e := m.makeBaseRule(rule)
	if e == nil {
		newlog.Info("add new rule, ruleId=%d, name=%s, type=%d, range=%d", rule.ID, rule.Name, rule.Type, rule.QueryRange)
		m.alarmRules[rule.ID] = baseRule
		baseRule.Start()
	}
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
