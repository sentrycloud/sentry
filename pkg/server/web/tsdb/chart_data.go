package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

type ChartDataReq struct {
	dbmodel.Chart
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type ChartData struct {
	protocol.CurveData
	Name string `json:"name"`
}

func QueryChartData(w http.ResponseWriter, r *http.Request) {
	var chartDataReq ChartDataReq
	err := protocol.DecodeRequest(r, &chartDataReq)
	if err != nil {
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	downSample, err := protocol.TransferDownSample(chartDataReq.DownSample)
	if err != nil || downSample == 0 {
		newlog.Error("parse downSample failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeDownSampleError, nil)
		return
	}

	lines, err := dbmodel.QueryChatLines(chartDataReq.ID)
	if err != nil {
		newlog.Error("query mysql failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
		return
	}

	chartDataReq.Start *= 1000 // transfer to milliseconds
	chartDataReq.End *= 1000
	var curveDataList []ChartData
	for _, line := range lines {
		var tags = map[string]string{}
		err = protocol.Json.UnmarshalFromString(line.Tags, &tags)
		if err != nil {
			newlog.Error("unmarshal tags failed for charId=%d, lineId=%d, tags=%s", chartDataReq.ID, line.ID, line.Tags)
			continue
		}

		m := protocol.MetricReq{
			Metric: line.Metric,
			Tags:   tags,
		}

		offset := int64(line.Offset * 3600 * 24 * 1000)
		sql := buildRangeQuerySql(chartDataReq.Start+offset, chartDataReq.End+offset, chartDataReq.Aggregation, downSample, &m)
		results, e := QueryTSDB(sql, 2)
		if e != nil {
			newlog.Error("query TSDB failed: %v", e) // log the error, but still return success
			continue
		}

		var dataPoints []protocol.TimeValuePoint
		for _, row := range results {
			point := protocol.TimeValuePoint{
				TimeStamp: (row[0].(int64) - offset) / 1000, // add back offset, so data can be displayed on the same axis
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

	protocol.WriteQueryResp(w, protocol.CodeOK, curveDataList)
}
