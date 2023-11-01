package http

import (
	"database/sql/driver"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func queryTimeSeriesDataForRange(w http.ResponseWriter, r *http.Request) {
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

	conn, err := connPool.GetConn()
	if err != nil {
		newlog.Error("queryTagKeys: get conn pool failed: %v", err)
		resp.Code = CodeGetConnPoolError
		resp.Msg = CodeMsg[CodeGetConnPoolError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	defer connPool.PutConn(conn)

	var curveDataList []protocol.CurveData
	for _, m := range req.Metrics {
		sql := buildRangeQuerySql(req.Start, req.End, req.Aggregator, req.DownSample, &m)
		rows, err := conn.Query(sql)
		if err != nil {
			newlog.Error("queryTagKeys: query taos failed: %v", err)
			resp.Code = CodeExecSqlError
			resp.Msg = CodeMsg[CodeExecSqlError]
			writeQueryResp(w, http.StatusBadRequest, &resp)
			return
		}

		var dataPoints []protocol.TimeValuePoint
		for {
			values := make([]driver.Value, 2) // field, type, length, note
			if rows.Next(values) != nil {
				break
			}

			point := protocol.TimeValuePoint{
				TimeStamp: values[0].(int64) / 1000,
				Value:     values[1].(float64),
			}

			dataPoints = append(dataPoints, point)
		}
		_ = rows.Close()

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
