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

type TopNRule struct {
	AlarmRule
}

func (r *TopNRule) Parse() error {
	err := r.AlarmRule.Parse()
	if err != nil {
		return err
	}

	starCount := 0
	for _, v := range r.alarmDataSource.Tags {
		if strings.Contains(v, "*") {
			starCount++
		}
	}

	if starCount != 1 {
		newlog.Error("TopN alarm rule must has one and only one * tag for ruleId=%d", r.Id)
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
		Aggregator: r.alarmDataSource.Aggregator,
		DownSample: int64(r.alarmDataSource.DownSample),
		Order:      r.alarmDataSource.Sort,
		Limit:      r.alarmDataSource.Limit,
	}

	topNList, err := query.TopN(request)
	if err != nil {
		newlog.Error("query topN data failed for ruleId=%d", r.Id)
		return
	}

	for _, topNData := range topNList {
		if topNData.Value < r.alarmTrigger.LessThan || topNData.Value > r.alarmTrigger.GreaterThan {
			alarmMessage := r.buildAlarmMessage(topNData.Metric, topNData.Tags, startTime, topNData.Value)
			sender.WeChatMessage(alarmMessage)
		}
	}
}
