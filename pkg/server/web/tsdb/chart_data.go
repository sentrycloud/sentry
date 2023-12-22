package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
	"time"
)

type ChartDataReq struct {
	Start   int64  `json:"start"`
	End     int64  `json:"end"`
	ChartId uint32 `json:"chart_id"`
}

type ChartData struct {
	protocol.CurveData
	Name string `json:"name"`
}

func QueryChartData(w http.ResponseWriter, r *http.Request) {
	var chartDataReq ChartDataReq
	err := protocol.Json.NewDecoder(r.Body).Decode(&chartDataReq)
	if err != nil {
		errMsg := "json decode failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 1, errMsg, nil)
		return
	}

	lines, err := dbmodel.QueryChatLines(chartDataReq.ChartId)
	if err != nil {
		errMsg := "query mysql failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 2, errMsg, nil)
		return
	}

	var curveDataList []ChartData
	for _, line := range lines {
		var tags = map[string]string{}
		protocol.Json.UnmarshalFromString(line.Tags, &tags)
		m := protocol.MetricReq{
			Metric: line.Metric,
			Tags:   tags,
		}

		now := time.Now().UnixMilli()
		start := now - 1800*1000
		sql := buildRangeQuerySql(start, now, "avg", 10, &m)
		results, e := QueryTSDB(sql, 2)
		if e != nil {
			continue
		}

		var dataPoints []protocol.TimeValuePoint
		for _, row := range results {
			point := protocol.TimeValuePoint{
				TimeStamp: row[0].(int64) / 1000,
				Value:     row[1].(float64),
			}

			dataPoints = append(dataPoints, point)
		}

		curveData := ChartData{
			Name: line.Name,
		}
		curveData.Metric = line.Metric
		curveData.Tags = tags
		curveData.DPS = dataPoints

		curveDataList = append(curveDataList, curveData)
	}

	var resp = protocol.QueryResp{}
	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = curveDataList
	writeQueryResp(w, http.StatusOK, &resp)

}
