package mysql

import (
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

}

func deleteChart(w http.ResponseWriter, r *http.Request) {

}

func validateChartParam(params *ChartParams) error {
	return nil
}
