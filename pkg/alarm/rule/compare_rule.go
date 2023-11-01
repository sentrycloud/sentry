package rule

import (
	"errors"
	"github.com/sentrycloud/sentry/pkg/alarm/query"
	"github.com/sentrycloud/sentry/pkg/alarm/schedule"
	"github.com/sentrycloud/sentry/pkg/alarm/sender"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"sort"
	"strconv"
	"strings"
	"time"
)

const MaxCompareDays = 7

type CompareRule struct {
	AlarmRule

	hasStarTags         bool
	lastTimeUpdateCurve int64
	curves              []protocol.MetricReq
}

func (r *CompareRule) Parse() error {
	err := r.AlarmRule.Parse()
	if err != nil {
		return err
	}

	if r.alarmDataSource.CompareType != CompareTypeDifference && r.alarmDataSource.CompareType != CompareTypeRatio {
		newlog.Error("compare type is invalid for ruleId=%d", r.Id)
		return errors.New("compare type is invalid")
	}

	if r.alarmDataSource.CompareDaysAgo > MaxCompareDays {
		newlog.Error("compare day ago is too large for ruleId=%d", r.Id)
		return errors.New("compare day ago is too large")
	}

	r.hasStarTags, r.curves, err = r.parseTags()
	return err
}

func (r *CompareRule) Start() {
	r.alarmTimer = schedule.Repeat(r.alarmInterval, r.Run)
}

func (r *CompareRule) Run() {
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

	// query history data list
	rangeReq.Start = startTime - int64(r.alarmDataSource.CompareDaysAgo*3600*24+r.alarmDataSource.CompareSeconds)
	rangeReq.End = rangeReq.Start + int64(r.alarmDataSource.CompareSeconds)
	curveHistoryDataList, err := query.Range(rangeReq)
	if err != nil {
		newlog.Error("query history range data failed for ruleId=%d", r.Id)
		return
	}

	// calculate history data average value for each curve
	var historyValueMap = make(map[string]float64)
	for _, curveData := range curveHistoryDataList {
		dpsCount := len(curveData.DPS)
		if dpsCount > 0 {
			key := tags2Str(curveData.Tags)
			var values float64 = 0
			for _, dp := range curveData.DPS {
				values += dp.Value
			}
			historyValueMap[key] = values / float64(dpsCount)
		}
	}

	for _, curveData := range curveDataList {
		tags := tags2Str(curveData.Tags)
		historyValue, exist := historyValueMap[tags]
		if !exist {
			continue
		}

		triggerCount := 0
		var triggerTime int64
		var triggerValue float64
		var triggerCompareValue float64
		for _, dp := range curveData.DPS {
			var compareValue float64
			if r.alarmDataSource.CompareType == CompareTypeDifference {
				compareValue = dp.Value - historyValue
			} else {
				compareValue = dp.Value / historyValue // golang divide by zero will get inf, and it's larger than math.MaxFloat64
			}

			if compareValue < r.alarmTrigger.LessThan || compareValue > r.alarmTrigger.GreaterThan {
				triggerCount++
				triggerTime = dp.TimeStamp
				triggerValue = dp.Value
				triggerCompareValue = compareValue
			}
		}

		if triggerCount >= r.alarmTrigger.ErrorCount {
			alarmMessage := r.buildAlarmMessage(curveData.Metric, curveData.Tags, triggerTime, triggerValue)
			alarmMessage = r.buildCompareAlarmMessage(alarmMessage, historyValue, triggerCompareValue)
			sender.WeChatMessage(alarmMessage)
		}
	}
}

// replace placeholder in compare alarm message template to actual value, example placeholder:
// {history.value}, {compare.value}
func (r *CompareRule) buildCompareAlarmMessage(result string, historyValue float64, compareValue float64) string {
	var searchList []string
	var replaceList []string

	searchList = append(searchList, "{history.value}")
	replaceList = append(replaceList, strconv.FormatFloat(historyValue, 'f', -1, 64))

	searchList = append(searchList, "{compare.value}")
	replaceList = append(replaceList, strconv.FormatFloat(compareValue, 'f', -1, 64))

	for index := range replaceList {
		result = strings.ReplaceAll(result, searchList[index], replaceList[index])
	}

	return result
}

func tags2Str(tags map[string]string) string {
	tagsStr := ""
	if len(tags) == 0 {
		return tagsStr
	}
	var tmpSlice []string
	for k, v := range tags {
		tmpSlice = append(tmpSlice, k+":"+v)
	}
	sort.Strings(tmpSlice)
	return strings.Join(tmpSlice, ",")
}
