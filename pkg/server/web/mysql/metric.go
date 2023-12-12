package mysql

import (
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func HandleMetricWhiteList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		queryMetricWhiteList(w, r)
	case "PUT":
		addMetricWhiteList(w, r)
	case "POST":
		updateMetricWhiteList(w, r)
	case "DELETE":
		deleteMetricWhiteList(w, r)
	default:
		protocol.MethodNotSupport(w, r)
	}
}

func queryMetricWhiteList(w http.ResponseWriter, r *http.Request) {
	contacts, err := dbmodel.QueryAllMetricWhiteList()
	if err != nil {
		newlog.Error("db query failed: %v", err)
		protocol.WriteQueryResp(w, http.StatusInternalServerError, 1, "db query failed", nil)
	} else {
		protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", contacts)
	}
}

func addMetricWhiteList(w http.ResponseWriter, r *http.Request) {
	modifyMetricWhiteList(w, r, dbmodel.AddMetricWhiteList)
}

func updateMetricWhiteList(w http.ResponseWriter, r *http.Request) {
	modifyMetricWhiteList(w, r, dbmodel.UpdateMetricWhiteList)
}

func deleteMetricWhiteList(w http.ResponseWriter, r *http.Request) {
	modifyMetricWhiteList(w, r, dbmodel.DeleteMetricWhiteList)
}

func modifyMetricWhiteList(w http.ResponseWriter, r *http.Request, modifyFunc func(contact *dbmodel.MetricWhiteList) error) {
	var metric dbmodel.MetricWhiteList
	err := protocol.Json.NewDecoder(r.Body).Decode(&metric)
	if err != nil {
		newlog.Error("json decode failed: %v", err)
		protocol.WriteQueryResp(w, http.StatusInternalServerError, 2, "json decoder failed", nil)
		return
	}

	err = modifyFunc(&metric)
	if err != nil {
		newlog.Error("db modify failed: %v", err)
		protocol.WriteQueryResp(w, http.StatusInternalServerError, 3, "db modify failed", nil)
		return
	}

	protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", metric)
}
