package mysql

import (
	"errors"
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
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
		protocol.MethodNotSupport(w, r)
	}
}

func HandleChartList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		protocol.MethodNotSupport(w, r)
		return
	}

	// request body format: {"dashboard_id": xxx}
	var chart dbmodel.Chart
	err := protocol.Json.NewDecoder(r.Body).Decode(&chart)
	if err != nil {
		errMsg := "json decode failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 1, errMsg, nil)
		return
	}

	chartList, err := dbmodel.QueryDashboardCharts(chart.DashboardId)
	if err != nil {
		errMsg := "query chart list failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 2, errMsg, nil)
		return
	}

	protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", chartList)
}

func getChart(w http.ResponseWriter, r *http.Request) {
	var chart dbmodel.Chart
	err := protocol.Json.NewDecoder(r.Body).Decode(&chart)
	if err != nil {
		errMsg := "json decode failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 1, errMsg, nil)
		return
	}

	err = dbmodel.GetEntity(&chart)
	if err != nil {
		errMsg := "query chart entity failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 1, errMsg, nil)
		return
	}

	lines, err := dbmodel.QueryChatLines(chart.ID)
	if err != nil {
		errMsg := "query chart lines failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 1, errMsg, nil)
		return
	}

	var chartParams ChartParams
	chartParams.Chart = chart
	chartParams.Lines = lines

	protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", &chartParams)
}

func addChart(w http.ResponseWriter, r *http.Request) {
	var chartParams ChartParams
	err := protocol.Json.NewDecoder(r.Body).Decode(&chartParams)
	if err != nil {
		errMsg := "json decode failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 1, errMsg, nil)
		return
	}

	err = validateChartParam(&chartParams)
	if err != nil {
		errMsg := "validate params failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 2, errMsg, nil)
		return
	}

	// no transaction here, may have partial failure
	// add chart
	err = dbmodel.AddEntity(&chartParams.Chart)
	if err != nil {
		errMsg := "add chart to db failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 3, errMsg, nil)
		return
	}

	// add lines
	for _, line := range chartParams.Lines {
		line.ChartId = chartParams.Chart.ID // update just insert chartId for all lines
		err = dbmodel.AddEntity(&line)
		if err != nil {
			errMsg := "add line to db failed: " + err.Error()
			newlog.Error(errMsg)
			protocol.WriteQueryResp(w, http.StatusOK, 4, errMsg, nil)
			return
		}
	}

	protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", nil)
}

func modifyChart(w http.ResponseWriter, r *http.Request) {
	var chartParams ChartParams
	err := protocol.Json.NewDecoder(r.Body).Decode(&chartParams)
	if err != nil {
		errMsg := "json decode failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 1, errMsg, nil)
		return
	}

	err = validateChartParam(&chartParams)
	if err != nil {
		errMsg := "validate params failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 2, errMsg, nil)
		return
	}

	err = dbmodel.UpdateEntity(&chartParams.Chart)
	if err != nil {
		errMsg := "update db failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 3, errMsg, nil)
		return
	}

	oldLines, err := dbmodel.QueryChatLines(chartParams.ID)
	if err != nil {
		errMsg := "query chart lines failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 4, errMsg, nil)
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
			errMsg := "update chart line failed: " + err.Error()
			newlog.Error(errMsg)
			protocol.WriteQueryResp(w, http.StatusOK, 5, errMsg, nil)
			return
		}
	}

	for _, line := range addLines {
		err = dbmodel.AddEntity(&line)
		if err != nil {
			errMsg := "add chart line failed: " + err.Error()
			newlog.Error(errMsg)
			protocol.WriteQueryResp(w, http.StatusOK, 6, errMsg, nil)
			return
		}
	}

	err = dbmodel.DeleteLines(deleteLineIds)
	if err != nil {
		errMsg := "delete chart line failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 7, errMsg, nil)
		return
	}

	protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", nil)
}

func deleteChart(w http.ResponseWriter, r *http.Request) {
	var chart dbmodel.Chart
	err := protocol.Json.NewDecoder(r.Body).Decode(&chart)
	if err != nil {
		errMsg := "json decode failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 1, errMsg, nil)
		return
	}

	newlog.Info("delete chart: dashboardId=%d, chartId=%d", chart.DashboardId, chart.ID)
	err = dbmodel.DeleteChartAndLines(chart.ID)
	if err != nil {
		errMsg := "delete in db failed: " + err.Error()
		newlog.Error(errMsg)
		protocol.WriteQueryResp(w, http.StatusOK, 2, errMsg, nil)
		return
	}

	protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", nil)
}

func validateChartParam(chart *ChartParams) error {
	if len(chart.Name) == 0 {
		return errors.New("chart name is empty")
	}

	if chart.Type != "line" && chart.Type != "pie" && chart.Type != "bar" {
		return errors.New("chart type not in line/pie/bar")
	}

	if chart.Aggregation != "sum" && chart.Aggregation != "avg" && chart.Aggregation != "max" && chart.Aggregation != "min" {
		return errors.New("chart aggregation not in sum/avg/max/min")
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
		err := protocol.Json.UnmarshalFromString(line.Tags, &tags)
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
