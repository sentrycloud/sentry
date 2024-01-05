package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func internalQueryRange(start int64, end int64, offset int64, aggregator string, downSample int64,
	metricReq *protocol.MetricReq) (*protocol.CurveData, int) {
	sql := buildRangeQuerySql(start+offset, end+offset, aggregator, downSample, metricReq)
	results, e := QueryTSDB(sql, 2)
	if e != nil {
		return nil, protocol.CodeExecTSDBSqlError
	}

	var dataPoints []protocol.TimeValuePoint
	for _, row := range results {
		point := protocol.TimeValuePoint{
			TimeStamp: (row[0].(int64) - offset) / 1000, // add back offset, so multiple line can be displayed on the same axis
			Value:     row[1].(float64),
		}

		dataPoints = append(dataPoints, point)
	}

	curveData := protocol.CurveData{
		Metric: metricReq.Metric,
		Tags:   metricReq.Tags,
		DPS:    dataPoints,
	}

	return &curveData, protocol.CodeOK
}

func QueryTimeSeriesDataForRange(w http.ResponseWriter, r *http.Request) {
	var req protocol.TimeSeriesDataRequest
	err := protocol.DecodeRequest(r, &req)
	if err != nil {
		newlog.Error("queryTimeSeriesDataForRange: decode query request failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	code := transferTimeSeriesDataRequest(&req)
	if code != protocol.CodeOK {
		newlog.Error("queryTimeSeriesDataForRange: transferTimeSeriesDataRequest failed: %s", protocol.CodeMsg[code])
		protocol.WriteQueryResp(w, code, nil)
		return
	}

	var curveDataList []*protocol.CurveData
	for _, m := range req.Metrics {
		curveData, retCode := internalQueryRange(req.Start, req.End, 0, req.Aggregator, req.DownSample, &m)
		if retCode != protocol.CodeOK {
			protocol.WriteQueryResp(w, retCode, nil)
			return
		}
		curveDataList = append(curveDataList, curveData)
	}

	protocol.WriteQueryResp(w, protocol.CodeOK, curveDataList)
}
