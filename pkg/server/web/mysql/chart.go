package mysql

import (
	"errors"
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
	"strconv"
)

type ChartParams struct {
	dbmodel.Chart
	Lines []dbmodel.Line `json:"lines"`
}

func HandleChart(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getChart(w, r)
	case "PUT":
		addChart(w, r)
	case "POST":
		modifyChart(w, r)
	case "DELETE":
		deleteChart(w, r)
	default:
		protocol.MethodNotSupport(w)
	}
}

func HandleChartList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		protocol.MethodNotSupport(w)
		return
	}

	// request body format: {"dashboard_id": xxx}
	var chart dbmodel.Chart
	err := protocol.DecodeRequest(r, &chart)
	if err != nil {
		newlog.Error("json decode failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	chartList, err := dbmodel.QueryDashboardCharts(chart.DashboardId)
	if err != nil {
		newlog.Error("query chart list failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
		return
	}

	protocol.WriteQueryResp(w, protocol.CodeOK, chartList)
}

func getChart(w http.ResponseWriter, r *http.Request) {
	chartId, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		newlog.Error("get chartId failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	var chart dbmodel.Chart
	chart.ID = uint32(chartId)
	err = dbmodel.GetEntity(&chart)
	if err != nil {
		newlog.Error("query chart entity failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
		return
	}

	lines, err := dbmodel.QueryChatLines(chart.ID)
	if err != nil {
		newlog.Error("query chart lines failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
		return
	}

	var chartParams ChartParams
	chartParams.Chart = chart
	chartParams.Lines = lines
	protocol.WriteQueryResp(w, protocol.CodeOK, &chartParams)
}

func addChart(w http.ResponseWriter, r *http.Request) {
	var chartParams ChartParams
	err := protocol.DecodeRequest(r, &chartParams)
	if err != nil {
		newlog.Error("json decode failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	err = validateChartParam(&chartParams)
	if err != nil {
		newlog.Error("validate params failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeInvalidParamError, nil)
		return
	}

	// no transaction here, may have partial failure
	// add chart
	err = dbmodel.AddEntity(&chartParams.Chart)
	if err != nil {
		newlog.Error("add chart to db failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
		return
	}

	// add lines
	for _, line := range chartParams.Lines {
		line.ChartId = chartParams.Chart.ID // update just insert chartId for all lines
		err = dbmodel.AddEntity(&line)
		if err != nil {
			newlog.Error("add line to db failed: %v", err)
			protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
			return
		}
	}

	protocol.WriteQueryResp(w, protocol.CodeOK, nil)
}

func modifyChart(w http.ResponseWriter, r *http.Request) {
	var chartParams ChartParams
	err := protocol.DecodeRequest(r, &chartParams)
	if err != nil {
		newlog.Error("json decode failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	err = validateChartParam(&chartParams)
	if err != nil {
		newlog.Error("validate params failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeInvalidParamError, nil)
		return
	}

	err = dbmodel.UpdateEntity(&chartParams.Chart)
	if err != nil {
		newlog.Error("update db failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
		return
	}

	oldLines, err := dbmodel.QueryChatLines(chartParams.ID)
	if err != nil {
		newlog.Error("query chart lines failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
		return
	}

	// calculate which lines need to be updated, added and deleted
	var updateLines []dbmodel.Line
	for _, oldLine := range oldLines {
		for _, newLine := range chartParams.Lines {
			if oldLine.ID == newLine.ID {
				updateLines = append(updateLines, newLine)
			}
		}
	}

	var addLines []dbmodel.Line
	for _, line := range chartParams.Lines {
		if !isInLines(updateLines, line.ID) {
			addLines = append(addLines, line)
		}
	}

	var deleteLineIds []uint32
	for _, line := range oldLines {
		if !isInLines(updateLines, line.ID) {
			deleteLineIds = append(deleteLineIds, line.ID)
		}
	}

	for _, line := range updateLines {
		err = dbmodel.UpdateEntity(&line)
		if err != nil {
			newlog.Error("update chart line failed: %v", err)
			protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
			return
		}
	}

	for _, line := range addLines {
		err = dbmodel.AddEntity(&line)
		if err != nil {
			newlog.Error("add chart line failed: %v", err)
			protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
			return
		}
	}

	if len(deleteLineIds) > 0 {
		err = dbmodel.DeleteLines(deleteLineIds)
		if err != nil {
			newlog.Error("delete chart line failed: %v", err)
			protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
			return
		}
	}

	protocol.WriteQueryResp(w, protocol.CodeOK, nil)
}

func deleteChart(w http.ResponseWriter, r *http.Request) {
	var chart dbmodel.Chart
	err := protocol.DecodeRequest(r, &chart)
	if err != nil {
		newlog.Error("json decode failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	newlog.Info("delete chart: dashboardId=%d, chartId=%d", chart.DashboardId, chart.ID)
	err = dbmodel.DeleteChartAndLines(chart.ID)
	if err != nil {
		newlog.Error("delete in db failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
		return
	}

	protocol.WriteQueryResp(w, protocol.CodeOK, nil)
}

func validateChartParam(chart *ChartParams) error {
	if len(chart.Name) == 0 {
		return errors.New("chart name is empty")
	}

	if chart.Type != "line" && chart.Type != "pie" && chart.Type != "topN" {
		return errors.New("chart type not in line/pie/topN")
	}

	if chart.Aggregation != "sum" && chart.Aggregation != "avg" && chart.Aggregation != "max" && chart.Aggregation != "min" {
		return errors.New("chart aggregation not in sum/avg/max/min")
	}

	downSample, err := protocol.TransferDownSample(chart.DownSample)
	if err != nil || downSample == 0 {
		return errors.New("wrong down sample format")
	}

	if len(chart.Lines) == 0 {
		return errors.New("no line in chart")
	}

	for _, line := range chart.Lines {
		if len(line.Name) == 0 {
			return errors.New("no name in line")
		}

		if len(line.Metric) == 0 {
			return errors.New("no metric in line")
		}

		var tags = make(map[string]string)
		err = protocol.Json.UnmarshalFromString(line.Tags, &tags)
		if err != nil {
			return errors.New("can not unmarshal tags: " + err.Error())
		}
	}

	return nil
}

func isInLines(lines []dbmodel.Line, lineId uint32) bool {
	for _, line := range lines {
		if line.ID == lineId {
			return true
		}
	}

	return false
}
