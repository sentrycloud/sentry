package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
	"strings"
)

const OneDayMilliseconds = 3600 * 24 * 1000
const LineCountLimit = 100 // limit the max line count that a query can take

type ChartDataReq struct {
	dbmodel.Chart
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type ChartData struct {
	*protocol.CurveData
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

	if chartDataReq.Type == "topN" {
		if len(lines) != 1 {
			newlog.Error("topN can have one line only")
			protocol.WriteQueryResp(w, protocol.CodeInvalidParamError, nil)
			return
		}

		var tags = map[string]string{}
		err = protocol.Json.UnmarshalFromString(lines[0].Tags, &tags)
		if err != nil {
			newlog.Error("unmarshal tags failed for charId=%d, lineId=%d, tags=%s", chartDataReq.ID, lines[0].ID, lines[0].Tags)
			protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
			return
		}

		req := protocol.TopNRequest{
			Start:      chartDataReq.Start,
			End:        chartDataReq.End,
			Aggregator: chartDataReq.Aggregation,
			DownSample: downSample,
			Limit:      chartDataReq.TopnLimit,
			Order:      "desc",
			Metric:     lines[0].Metric,
			Tags:       tags,
		}

		topNDataList, code := internalQueryTopN(&req)
		protocol.WriteQueryResp(w, code, topNDataList)
	} else {
		chartDataReq.Start *= 1000 // transfer to milliseconds
		chartDataReq.End *= 1000
		var chartDataList []ChartData
		for _, line := range lines {
			var tags = map[string]string{}
			err = protocol.Json.UnmarshalFromString(line.Tags, &tags)
			if err != nil {
				newlog.Error("unmarshal tags failed for charId=%d, lineId=%d, tags=%s", chartDataReq.ID, line.ID, line.Tags)
				continue
			}

			metricReq := protocol.MetricReq{
				Metric: line.Metric,
				Tags:   tags,
			}

			curveList, code := internalQueryCurves(&metricReq)
			if code != protocol.CodeOK || curveList == nil {
				continue
			}

			if len(curveList) > LineCountLimit {
				// do not query too many line once, that maybe hog too much resource
				curveList = curveList[:LineCountLimit]
			}

			for _, curve := range curveList {
				m := protocol.MetricReq{
					Metric: line.Metric,
					Tags:   curve,
				}
				offset := int64(line.Offset * OneDayMilliseconds)
				curveData, retCode := internalQueryRange(chartDataReq.Start, chartDataReq.End, offset, chartDataReq.Aggregation, downSample, &m)
				if retCode != protocol.CodeOK {
					continue // query error still return success
				}

				lineName := getLineName(line.Name, len(curveList), tags, curve)
				chartData := ChartData{
					CurveData: curveData,
					Name:      lineName,
				}

				chartDataList = append(chartDataList, chartData)
			}
		}
		protocol.WriteQueryResp(w, protocol.CodeOK, chartDataList)
	}
}

func getLineName(lineName string, curveCount int, tags, curve map[string]string) string {
	if curveCount <= 1 {
		return lineName
	}

	// tag value has *, use tagValue as name to differentiate each other
	lineName = ""
	for k, v := range curve {
		oldV := tags[k]
		if strings.HasSuffix(oldV, "*") {
			if len(lineName) > 0 {
				lineName += "-"
			}
			lineName += v
		}
	}
	return lineName
}
