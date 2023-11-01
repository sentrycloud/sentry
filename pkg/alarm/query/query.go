package query

import (
	"errors"
	"github.com/buger/jsonparser"
	"github.com/sentrycloud/sentry/pkg/httpclient"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
)

var serverAddr string

func InitServerAddr(addr string) {
	serverAddr = addr
}

func requestSentryServer(url string, request interface{}, response interface{}) error {
	content, err := protocol.Json.Marshal(request)
	if err != nil {
		newlog.Error("marshal http request content failed: %v", err)
		return err
	}

	resp, err := httpclient.Call("POST", serverAddr+url, content, nil)
	if err != nil {
		newlog.Error("http call failed: %v", err)
		return err
	}

	code, err := jsonparser.GetInt(resp, "code")
	if code != 0 {
		msg, _ := jsonparser.GetString(resp, "msg")
		newlog.Error("http response failed: %s", msg)
		return errors.New(msg)
	}

	data, _, _, err := jsonparser.Get(resp, "data")
	if err != nil {
		newlog.Error("json get data failed: %v", err)
		return err
	}

	err = protocol.Json.Unmarshal(data, response)
	if err != nil {
		newlog.Error("unmarshal curve data list failed: %v", err)
		return err
	}
	return nil
}

func Curve(request *protocol.MetricReq) ([]map[string]string, error) {
	var curveList []map[string]string
	err := requestSentryServer(protocol.CurveUrl, request, &curveList)
	return curveList, err
}

func Range(request *protocol.TimeSeriesDataRequest) ([]protocol.CurveData, error) {
	var curveDataList []protocol.CurveData
	err := requestSentryServer(protocol.RangeUrl, request, &curveDataList)
	return curveDataList, err
}

func TopN(request *protocol.TopNRequest) ([]protocol.TopNData, error) {
	var topnDataList []protocol.TopNData
	err := requestSentryServer(protocol.TopNUrl, request, &topnDataList)
	return topnDataList, err
}
