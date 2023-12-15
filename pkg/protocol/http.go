package protocol

import (
	"encoding/json"
	"errors"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"io"
	"net/http"
	"strings"
)

const (
	// MetricUrl API for TSDB
	MetricUrl   = "/server/api/metrics"
	TagKeyUrl   = "/server/api/tagKeys"
	TagValueUrl = "/server/api/tagValues"
	CurveUrl    = "/server/api/curves"
	RangeUrl    = "/server/api/range"
	TopNUrl     = "/server/api/topn"

	// ContactUrl API for MySQL
	ContactUrl         = "/server/api/contact"
	MetricWhiteListUrl = "/server/api/metricWhiteList"
	DashboardUrl       = "/server/api/dashboard"
	ChartUrl           = "/server/api/chart"

	PutMetricsUrl = "/server/api/putMetrics"
)

type MetricReq struct {
	Metric  string              `json:"metric"`
	Tags    map[string]string   `json:"tags"`
	Filters map[string][]string `json:"filters"` // tag values with ||
}

type TimeSeriesDataRequest struct {
	Token      string      `json:"token"`
	Start      int64       `json:"start"`
	End        int64       `json:"end"`
	Last       int64       `json:"last"`
	Aggregator string      `json:"aggregator"`
	DownSample int64       `json:"down_sample"`
	Metrics    []MetricReq `json:"metrics"`
}

type TimeValuePoint struct {
	TimeStamp int64   `json:"ts"`
	Value     float64 `json:"v"`
}

type CurveData struct {
	Metric string            `json:"metric"`
	Tags   map[string]string `json:"tags"`
	DPS    []TimeValuePoint  `json:"dps"`
}

type QueryResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type TopNRequest struct {
	Token      string              `json:"token"`
	Start      int64               `json:"start"`
	End        int64               `json:"end"`
	Last       int64               `json:"last"`
	Aggregator string              `json:"aggregator"`
	DownSample int64               `json:"down_sample"`
	Limit      int                 `json:"limit"`
	Order      string              `json:"order"` // desc/asc
	Metric     string              `json:"metric"`
	Tags       map[string]string   `json:"tags"`
	Filters    map[string][]string // extracted from tags with "||" in the value
	Field      string              // extracted from tags with value of "*"
}

type TopNData struct {
	Metric string            `json:"metric"`
	Tags   map[string]string `json:"tags"`
	Value  float64           `json:"value"`
}

func CheckAggregator(aggregator string) (string, error) {
	aggregator = strings.ToLower(aggregator)
	if aggregator == "sum" || aggregator == "avg" || aggregator == "max" || aggregator == "min" {
		return aggregator, nil
	}
	return aggregator, errors.New("no such aggregator: " + aggregator)
}

func CheckOrder(order string) (string, error) {
	if len(order) == 0 {
		return "desc", nil // default order is descendent
	}

	order = strings.ToLower(order)
	if order == "desc" || order == "asc" {
		return order, nil
	}

	return "", errors.New("no such order: " + order)
}

func CollectHttpMetrics(w http.ResponseWriter, req *http.Request) ([]MetricValue, error) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		newlog.Error("simple put read failed: %s", err)
		return nil, err
	}

	var values []MetricValue
	err = json.Unmarshal(data, &values)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		newlog.Error("json format invalid: %s", err)
		return nil, err
	}

	w.WriteHeader(http.StatusOK)
	return values, nil
}

func WriteQueryResp(w http.ResponseWriter, status int, code int, msg string, data interface{}) {
	var resp = &QueryResp{}
	resp.Code = code
	resp.Msg = msg
	resp.Data = data
	jsonData, err := Json.Marshal(resp)
	if err != nil {
		newlog.Error("marsh query response failed")
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	w.Write(jsonData)
}

func MethodNotSupport(w http.ResponseWriter, r *http.Request) {
	WriteQueryResp(w, http.StatusBadRequest, 1, "http method not support", nil)
}
