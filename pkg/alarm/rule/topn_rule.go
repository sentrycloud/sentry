package rule

import (
	"errors"
	"github.com/sentrycloud/sentry/pkg/alarm/query"
	"github.com/sentrycloud/sentry/pkg/alarm/schedule"
	"github.com/sentrycloud/sentry/pkg/alarm/sender"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"strings"
)

const MaxTopNLimit = 100

type TopNRule struct {
	AlarmRule
}

func (r *TopNRule) Parse() error {
	err := r.AlarmRule.Parse()
	if err != nil {
		return err
	}

	r.alarmDataSource.Sort, err = protocol.CheckOrder(r.alarmDataSource.Sort)
	if err != nil {
		newlog.Error("wrong order for ruleId=%d, err: %v", r.ID, err)
		return err
	}

	if r.alarmDataSource.Limit <= 0 || r.alarmDataSource.Limit > MaxTopNLimit {
		newlog.Warn("set topN limit to default value for ruleId=%d", r.ID)
		r.alarmDataSource.Limit = MaxTopNLimit
	}

	starCount := 0
	for _, v := range r.alarmDataSource.Tags {
		if strings.Contains(v, "*") {
			starCount++
		}
	}

	if starCount != 1 {
		newlog.Error("TopN alarm rule must has one and only one * tag for ruleId=%d", r.ID)
		return errors.New("TopN alarm rule must has one and only one * tag")
	}

	return nil
}

func (r *TopNRule) Start() {
	r.alarmTimer = schedule.Repeat(r.alarmInterval, r.Run)
}

func (r *TopNRule) Run() {
	startTime, endTime := r.startEndTime()

	var request = &protocol.TopNRequest{
		Token:      "",
		Start:      startTime,
		End:        endTime,
		Metric:     r.alarmDataSource.Metric,
		Tags:       r.alarmDataSource.Tags,
		Aggregator: r.alarmDataSource.Aggregation,
		DownSample: r.alarmDataSource.DownSample,
		Order:      r.alarmDataSource.Sort,
		Limit:      r.alarmDataSource.Limit,
	}

	topNList, err := query.TopN(request)
	if err != nil {
		newlog.Error("query topN data failed for ruleId=%d", r.ID)
		return
	}

	for _, topNData := range topNList {
		if topNData.Value < r.alarmTrigger.LessThan || topNData.Value > r.alarmTrigger.GreaterThan {
			alarmMessage := r.buildAlarmMessage(topNData.Metric, topNData.Tags, startTime, topNData.Value)
			sender.MailMessage(r.Contacts, alarmMessage)
		}
	}
}
