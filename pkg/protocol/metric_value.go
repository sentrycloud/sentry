package protocol

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"regexp"
	"time"
)

const MaxExpireTime = 3600

type MetricValue struct {
	Metric    string            `json:"metric"`
	Tags      map[string]string `json:"tags"`
	Timestamp uint64            `json:"timestamp"`
	Value     float64           `json:"value"`
}

var validNameRegExp = regexp.MustCompile("[a-zA-Z_][a-zA-Z_0-9]*")

func (m *MetricValue) IsValid() bool {
	if len(m.Metric) == 0 {
		newlog.Error("metric is empty")
		return false
	}

	if !validNameRegExp.MatchString(m.Metric) {
		newlog.Error("metric=%s is not valid", m.Metric)
	}

	for k, v := range m.Tags {
		if len(k) == 0 || len(v) == 0 {
			newlog.Error("metric=%s has empty tags", m.Metric)
			return false
		}

		if !validNameRegExp.MatchString(k) {
			newlog.Error("tag=%s of metric=%s is not valid", k, m.Metric)
		}
	}

	if time.Now().Unix()-int64(m.Timestamp) > MaxExpireTime {
		newlog.Error("metric=%s, timestamp=%ld is too old", m.Metric, m.Timestamp)
		return false
	}

	return true
}
