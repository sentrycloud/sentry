package rule

import (
	"encoding/json"
	"errors"
	"github.com/RussellLuo/timingwheel"
	"github.com/sentrycloud/sentry/pkg/alarm/query"
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	AlarmTypeHeartBeat = iota
	AlarmTypeThreshold
	AlarmTypeTopN
	AlarmTypeCompare
)

const (
	CompareTypeDifference = iota
	CompareTypeRatio
)

const (
	UpdateCurvesInterval = 600
	AlarmQueryDelay      = 10 // the end of query time range is 10 seconds ago
)

type BaseRule interface {
	GetAlarmRule() dbmodel.AlarmRule
	Parse() error
	Start()
	Stop()
	Run()
}

type AlarmDataSource struct {
	Metric         string            `json:"metric"`
	Tags           map[string]string `json:"tags"`
	DownSample     int64             `json:"down_sample"`
	Aggregation    string            `json:"aggregation"`
	Sort           string            `json:"sort"`             // use for topN rule only
	Limit          int               `json:"limit"`            // use for topN rule only
	CompareType    int               `json:"compare_type"`     // use for compare rule only
	CompareDaysAgo int               `json:"compare_days_ago"` // use for compare rule only
	CompareSeconds int               `json:"compare_seconds"`  // use for compare rule only
}

type AlarmTrigger struct {
	ErrorCount  int     `json:"error_count"`
	LessThan    float64 `json:"less_than"`
	GreaterThan float64 `json:"greater_than"`
}

type AlarmRule struct {
	dbmodel.AlarmRule

	alarmDataSource AlarmDataSource
	alarmTrigger    AlarmTrigger
	alarmInterval   time.Duration
	alarmQueryDelay int64
	alarmTimer      *timingwheel.Timer
}

func (r *AlarmRule) GetAlarmRule() dbmodel.AlarmRule {
	return r.AlarmRule
}

func (r *AlarmRule) Parse() error {
	if len(r.DataSource) == 0 {
		newlog.Error("empty DataSource for ruleId=%d", r.ID)
		return errors.New("empty DataSource")
	}

	if len(r.Trigger) == 0 {
		newlog.Error("empty Trigger for ruleId=%d", r.ID)
		return errors.New("empty Trigger")
	}

	err := json.Unmarshal([]byte(r.DataSource), &r.alarmDataSource)
	if err != nil {
		newlog.Error("unmarshal DataSource for ruleId=%d, failed: %v", r.ID, err)
		return err
	}

	r.alarmTrigger.LessThan = -math.MaxFloat64
	r.alarmTrigger.GreaterThan = math.MaxFloat64
	err = json.Unmarshal([]byte(r.Trigger), &r.alarmTrigger)
	if err != nil {
		newlog.Error("unmarshal Trigger for ruleId=%d failed: %v", r.ID, err)
		return err
	}

	if len(r.alarmDataSource.Metric) == 0 {
		newlog.Error("empty metric for ruleId=%d", r.ID)
		return errors.New("empty metric")
	}

	for k, v := range r.alarmDataSource.Tags {
		if len(k) == 0 || len(v) == 0 {
			newlog.Error("empty key or value in tags for ruleId=%d", r.ID)
			return errors.New("empty key or value in tags")
		}

		if strings.Contains(k, "*") {
			newlog.Error("tag key has * for ruleId=%d", r.ID)
			return errors.New("tag key has *")
		}
	}

	r.alarmDataSource.Aggregation, err = protocol.CheckAggregator(r.alarmDataSource.Aggregation)
	if err != nil {
		newlog.Error("no such aggregator=%s for ruleId=%d", r.alarmDataSource.Aggregation, r.ID)
		return err
	}

	if r.alarmDataSource.DownSample < 1 {
		newlog.Error("down sample is 0 for ruleId=%d", r.ID)
		return errors.New("down sample is 0")
	}

	r.alarmInterval = time.Duration(r.QueryRange/2) * time.Second
	r.alarmQueryDelay = AlarmQueryDelay
	return nil
}

func (r *AlarmRule) Stop() {
	r.alarmTimer.Stop()
}

// used for threshold and compare rule
func (r *AlarmRule) parseTags() (bool, []protocol.MetricReq, error) {
	hasStarTags := false
	hasSplitTags := false
	var splitKey string
	var splitValues []string
	for k, v := range r.alarmDataSource.Tags {
		if strings.HasSuffix(v, "*") {
			hasStarTags = true
		}

		if strings.Contains(v, "||") {
			if hasSplitTags {
				hasSplitTags = true
				splitKey = k
				splitValues = strings.Split(v, "||")
			} else {
				newlog.Error("has too much split tags in ruleId=%d", r.ID)
				return false, nil, errors.New("too much split tags")
			}
		}
	}

	var curves []protocol.MetricReq
	if !hasStarTags {
		curves = r.initCurves(hasSplitTags, splitKey, splitValues)
	}

	return hasStarTags, curves, nil
}

func (r *AlarmRule) initCurves(hasSplitTags bool, splitKey string, splitValues []string) []protocol.MetricReq {
	var curves []protocol.MetricReq
	if !hasSplitTags {
		curve := protocol.MetricReq{
			Metric: r.alarmDataSource.Metric,
			Tags:   r.alarmDataSource.Tags,
		}
		curves = append(curves, curve)
	} else {
		for _, v := range splitValues {
			v = strings.TrimSpace(v)
			curve := protocol.MetricReq{
				Metric: r.alarmDataSource.Metric,
				Tags:   r.alarmDataSource.Tags,
			}
			curve.Tags[splitKey] = v
			curves = append(curves, curve)
		}
	}

	return curves
}

func (r *AlarmRule) updateCurves() ([]protocol.MetricReq, error) {
	request := &protocol.MetricReq{
		Metric: r.alarmDataSource.Metric,
		Tags:   r.alarmDataSource.Tags,
	}

	tagsList, err := query.Curve(request)
	if err != nil {
		newlog.Error("query Curve request failed: %v", err)
		return nil, err
	}

	var curves []protocol.MetricReq
	for _, tags := range tagsList {
		curve := protocol.MetricReq{
			Metric: r.alarmDataSource.Metric,
			Tags:   tags,
		}

		curves = append(curves, curve)
	}

	return curves, nil
}

func (r *AlarmRule) startEndTime() (int64, int64) {
	endTime := time.Now().Unix() - r.alarmQueryDelay
	startTime := endTime - int64(r.QueryRange)
	return startTime, endTime
}

// replace placeholder in alarm message template to actual value, example placeholder:
// {metric}, {tags}, {tag.xxx}, {time}, {value}
func (r *AlarmRule) buildAlarmMessage(metric string, tags map[string]string, ts int64, value float64) string {
	var searchList []string
	var replaceList []string

	searchList = append(searchList, "{metric}")
	replaceList = append(replaceList, metric)

	tagData, err := protocol.Json.Marshal(tags)
	if err == nil {
		searchList = append(searchList, "{tags}")
		replaceList = append(replaceList, string(tagData))
	}

	for k, v := range tags {
		searchList = append(searchList, "{tag."+k+"}")
		replaceList = append(replaceList, v)
	}

	searchList = append(searchList, "{time}")
	replaceList = append(replaceList, time.Unix(ts, 0).Format("2006-01-02 15:04:05"))

	searchList = append(searchList, "{value}")
	replaceList = append(replaceList, strconv.FormatFloat(value, 'f', -1, 64))

	var result = r.Message
	for index := range replaceList {
		result = strings.ReplaceAll(result, searchList[index], replaceList[index])
	}

	return result
}
