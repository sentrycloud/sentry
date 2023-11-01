package rule

import (
	"github.com/sentrycloud/sentry/pkg/alarm/query"
	"github.com/sentrycloud/sentry/pkg/alarm/schedule"
	"github.com/sentrycloud/sentry/pkg/alarm/sender"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"time"
)

type ThresholdRule struct {
	AlarmRule

	hasStarTags         bool
	lastTimeUpdateCurve int64
	curves              []protocol.MetricReq
}

func (r *ThresholdRule) Parse() error {
	err := r.AlarmRule.Parse()
	if err != nil {
		return err
	}

	r.hasStarTags, r.curves, err = r.parseTags()
	return err
}

func (r *ThresholdRule) Start() {
	r.alarmTimer = schedule.Repeat(r.alarmInterval, r.Run)
}

func (r *ThresholdRule) Run() {
	now := time.Now().Unix()
	if r.hasStarTags && (r.lastTimeUpdateCurve < now-UpdateCurvesInterval) {
		r.lastTimeUpdateCurve = now
		curves, _ := r.updateCurves()
		if curves != nil {
			r.curves = curves
		}
	}

	if r.curves == nil {
		newlog.Error("curves is still empty for ruleId=%d", r.Id)
		return
	}

	startTime, endTime := r.startEndTime()
	rangeReq := &protocol.TimeSeriesDataRequest{
		Token:      "",
		Start:      startTime,
		End:        endTime,
		Aggregator: r.alarmDataSource.Aggregator,
		DownSample: r.alarmDataSource.DownSample,
		Metrics:    r.curves,
	}

	curveDataList, err := query.Range(rangeReq)
	if err != nil {
		newlog.Error("query range data failed for ruleId=%d", r.Id)
		return
	}

	for _, curveData := range curveDataList {
		triggerCount := 0
		var triggerTime int64
		var triggerValue float64
		for _, dp := range curveData.DPS {
			if dp.Value < r.alarmTrigger.LessThan || dp.Value > r.alarmTrigger.GreaterThan {
				triggerCount++
				triggerTime = dp.TimeStamp
				triggerValue = dp.Value
			}
		}

		if triggerCount >= r.alarmTrigger.ErrorCount {
			alarmMessage := r.buildAlarmMessage(curveData.Metric, curveData.Tags, triggerTime, triggerValue)
			sender.WeChatMessage(alarmMessage)
		}
	}
}
