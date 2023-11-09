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

type HeartBeatRule struct {
	AlarmRule

	rangeRequest *protocol.TimeSeriesDataRequest
}

func (r *HeartBeatRule) Parse() error {
	err := r.AlarmRule.Parse()
	if err != nil {
		return err
	}

	for _, v := range r.alarmDataSource.Tags {
		if strings.Contains(v, "*") {
			newlog.Error("heartbeat alarm rule has * tag for ruleId=%d", r.Id)
			return errors.New("heartbeat alarm rule has * tag")
		}
	}

	metric := protocol.MetricReq{
		Metric: r.alarmDataSource.Metric,
		Tags:   r.alarmDataSource.Tags,
	}

	r.rangeRequest = &protocol.TimeSeriesDataRequest{
		Token:      "",
		Aggregator: r.alarmDataSource.Aggregator,
		DownSample: r.alarmDataSource.DownSample,
	}

	r.rangeRequest.Metrics = append(r.rangeRequest.Metrics, metric)
	return nil
}

func (r *HeartBeatRule) Start() {
	r.alarmTimer = schedule.Repeat(r.alarmInterval, r.Run)
}

func (r *HeartBeatRule) Run() {
	r.rangeRequest.Start, r.rangeRequest.End = r.startEndTime()

	curveDataList, err := query.Range(r.rangeRequest)
	if err != nil {
		newlog.Error("query range data failed for heartbeat ruleId=%d", r.Id)
		return
	}

	for _, curveData := range curveDataList {
		if len(curveData.DPS) == 0 {
			alarmMessage := r.buildAlarmMessage(r.alarmDataSource.Metric, r.alarmDataSource.Tags, r.rangeRequest.Start, 0)
			sender.MailMessage(r.Contacts, alarmMessage)
		}
	}
}
