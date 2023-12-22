package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func QueryTimeSeriesDataForRange(w http.ResponseWriter, r *http.Request) {
	var req protocol.TimeSeriesDataRequest
	var resp = protocol.QueryResp{}
	err := protocol.Json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		newlog.Error("queryTimeSeriesDataForRange: decode query request failed: %v", err)
		resp.Code = CodeJsonDecodeError
		resp.Msg = CodeMsg[CodeJsonDecodeError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	code := transferTimeSeriesDataRequest(&req)
	if code != CodeOK {
		newlog.Error("queryTimeSeriesDataForRange: transferTimeSeriesDataRequest failed: %s", CodeMsg[code])
		resp.Code = code
		resp.Msg = CodeMsg[code]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	var curveDataList []protocol.CurveData
	for _, m := range req.Metrics {
		sql := buildRangeQuerySql(req.Start, req.End, req.Aggregator, req.DownSample, &m)
		results, e := QueryTSDB(sql, 2)
		if e != nil {
			resp.Code = CodeExecSqlError
			resp.Msg = CodeMsg[CodeExecSqlError]
			WriteQueryResp(w, http.StatusOK, &resp)
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

	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = curveDataList
	writeQueryResp(w, http.StatusOK, &resp)
}
