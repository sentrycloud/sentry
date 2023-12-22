package tsdb

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

// QueryMetrics all metrics that start with the name in the request
func QueryMetrics(w http.ResponseWriter, r *http.Request) {
	var m protocol.MetricReq
	var resp = protocol.QueryResp{}
	err := protocol.Json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		newlog.Error("queryMetrics: decode query request failed: %v", err)
		resp.Code = CodeJsonDecodeError
		resp.Msg = CodeMsg[CodeJsonDecodeError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	sql := fmt.Sprintf("show stables like '%%%s%%'", m.Metric)
	results, err := QueryTSDB(sql, 1)
	if err != nil {
		resp.Code = CodeExecSqlError
		resp.Msg = CodeMsg[CodeExecSqlError]
		WriteQueryResp(w, http.StatusOK, &resp)
		return
	}

	var metrics []string
	for _, row := range results {
		if row[0] != nil {
			metrics = append(metrics, row[0].(string))
		}
	}

	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = metrics
	writeQueryResp(w, http.StatusOK, &resp)
}
