package query

import (
	"github.com/sentrycloud/sentry/pkg/protocol"
	"testing"
)

const (
	testServerAddr = "http://127.0.0.1:51001"
	startTime      = 1693884000
	endTime        = 1693884600
)

func TestCurve(t *testing.T) {
	InitServerAddr(testServerAddr)

	request := &protocol.MetricReq{
		Metric: "sentry_test_metric",
	}
	request.Tags = map[string]string{
		"ip":      "127.0.0.1",
		"machine": "*",
	}

	tagList, err := Curve(request)
	if err != nil {
		t.Errorf("Curve request failed: %v", err)
	}

	for _, tags := range tagList {
		t.Log(tags)
	}
}

func TestRange(t *testing.T) {
	InitServerAddr(testServerAddr)

	metric := protocol.MetricReq{
		Metric: "sentry_sys_cpu_usage",
	}

	request := &protocol.TimeSeriesDataRequest{
		Token:      "",
		Start:      startTime,
		End:        endTime,
		Aggregator: "avg",
		DownSample: 10,
	}
	request.Metrics = append(request.Metrics, metric)

	curveDataList, err := Range(request)
	if err != nil {
		t.Errorf("Range request failed: %v", err)
	}

	for _, curveData := range curveDataList {
		if curveData.Metric != "sentry_sys_cpu_usage" {
			t.Errorf("metric not math")
		}
	}
}

func TestTopN(t *testing.T) {
	InitServerAddr(testServerAddr)

	var request = &protocol.TopNRequest{
		Token:  "",
		Start:  startTime,
		End:    endTime,
		Metric: "sentry_sys_cpu_usage",
		Tags: map[string]string{
			"ip": "*",
		},
		Aggregator: "avg",
		DownSample: 10,
		Order:      "desc",
		Limit:      10,
	}

	topNList, err := TopN(request)
	if err != nil {
		t.Errorf("TopN request failed: %v", err)
	}

	for _, topNData := range topNList {
		if topNData.Metric != "sentry_sys_cpu_usage" {
			t.Errorf("metric not math")
		}
	}
}
