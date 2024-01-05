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
	MetricUrl    = "/server/api/metrics"
	TagKeyUrl    = "/server/api/tagKeys"
	TagValueUrl  = "/server/api/tagValues"
	CurveUrl     = "/server/api/curves"
	RangeUrl     = "/server/api/range"
	TopNUrl      = "/server/api/topn"
	ChartDataUrl = "/server/api/chartData"

	// AlarmRuleUrl API for MySQL
	AlarmRuleUrl       = "/server/api/alarmRule"
	ContactUrl         = "/server/api/contact"
	MetricWhiteListUrl = "/server/api/metricWhiteList"
	DashboardUrl       = "/server/api/dashboard"
	ChartUrl           = "/server/api/chart"
	ChartListUrl       = "/server/api/chartList"

	PutMetricsUrl = "/server/api/putMetrics"
)

const (
	CodeOK                 = 0
	CodeApiNotFound        = 1
	CodeMethodNotFound     = 2
	CodeJsonDecodeError    = 3
	CodeInvalidParamError  = 4
	CodeGetConnPoolError   = 5
	CodeExecTSDBSqlError   = 6
	CodeSplitTagsError     = 7
	CodeStarKeysError      = 8
	CodeMaxQueryRangeError = 9
	CodeAggregatorError    = 10
	CodeDownSampleError    = 11
	CodeMetricError        = 12
	CodeTagCountError      = 13
	CodeOrderError         = 14
	CodeExecMySQLError     = 15
)

var CodeMsg = map[int]string{
	CodeOK:                 "ok",
	CodeApiNotFound:        "api not found",
	CodeMethodNotFound:     "http method not found",
	CodeJsonDecodeError:    "json decode error",
	CodeInvalidParamError:  "invalid parameter error",
	CodeGetConnPoolError:   "get conn pool error",
	CodeExecTSDBSqlError:   "TSDB SQL execution error",
	CodeSplitTagsError:     "split tags error",
	CodeStarKeysError:      "star keys error",
	CodeMaxQueryRangeError: "max query range error",
	CodeAggregatorError:    "aggregator error",
	CodeDownSampleError:    "down sample error",
	CodeMetricError:        "metric error",
	CodeTagCountError:      "too many tag count error",
	CodeOrderError:         "no such order for topN query",
	CodeExecMySQLError:     "MySQL execution error",
}

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

// DecodeRequest decode and write err log in one place
func DecodeRequest(r *http.Request, entity interface{}) error {
	err := Json.NewDecoder(r.Body).Decode(entity)
	if err != nil {
		newlog.Error("json decode failed: %v", err)
	}
	return err
}

func WriteQueryResp(w http.ResponseWriter, code int, data interface{}) {
	httpStatus := http.StatusOK
	msg, exist := CodeMsg[code]
	if !exist {
		msg = "unknown error"
	}
	var resp = &QueryResp{}
	resp.Code = code
	resp.Msg = msg
	resp.Data = data
	jsonData, err := Json.Marshal(resp)
	if err != nil {
		newlog.Error("marsh query response failed: %v", err)
		httpStatus = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(httpStatus)
	w.Write(jsonData)
}

func MethodNotSupport(w http.ResponseWriter) {
	WriteQueryResp(w, CodeMethodNotFound, nil)
}
