package tsdb

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

// QueryTagKeys query all tags of a metric
func QueryTagKeys(w http.ResponseWriter, r *http.Request) {
	var m protocol.MetricReq
	var resp = protocol.QueryResp{}
	err := protocol.Json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		newlog.Error("queryTagKeys: decode query request failed: %v", err)
		resp.Code = CodeJsonDecodeError
		resp.Msg = CodeMsg[CodeJsonDecodeError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	sql := fmt.Sprintf("desc `%s`", m.Metric)
	results, err := QueryTSDB(sql, 4)
	if err != nil {
		resp.Code = CodeExecSqlError
		resp.Msg = CodeMsg[CodeExecSqlError]
		WriteQueryResp(w, http.StatusOK, &resp)
		return
	}

	var tags []string
	for _, row := range results {
		note := row[3].(string)
		if note == "TAG" {
			tags = append(tags, row[0].(string))
		}
	}

	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = tags
	writeQueryResp(w, http.StatusOK, &resp)
}
