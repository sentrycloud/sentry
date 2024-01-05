package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

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

	var curveDataList []protocol.CurveData
	for _, m := range req.Metrics {
		sql := buildRangeQuerySql(req.Start, req.End, req.Aggregator, req.DownSample, &m)
		results, e := QueryTSDB(sql, 2)
		if e != nil {
			protocol.WriteQueryResp(w, protocol.CodeExecTSDBSqlError, nil)
			return
		}

		var dataPoints []protocol.TimeValuePoint
		for _, row := range results {
			point := protocol.TimeValuePoint{
				TimeStamp: row[0].(int64) / 1000,
				Value:     row[1].(float64),
			}

			dataPoints = append(dataPoints, point)
		}

		curveData := protocol.CurveData{
			Metric: m.Metric,
			Tags:   m.Tags,
			DPS:    dataPoints,
		}
		curveDataList = append(curveDataList, curveData)
	}

	protocol.WriteQueryResp(w, protocol.CodeOK, curveDataList)
}
