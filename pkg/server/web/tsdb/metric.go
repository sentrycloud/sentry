package tsdb

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

// QueryMetrics all metrics that start with the name in the request
func QueryMetrics(w http.ResponseWriter, r *http.Request) {
	var m protocol.MetricReq
	err := protocol.DecodeRequest(r, &m)
	if err != nil {
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	sql := fmt.Sprintf("show stables like '%%%s%%'", m.Metric)
	results, err := QueryTSDB(sql, 1)
	if err != nil {
		protocol.WriteQueryResp(w, protocol.CodeExecTSDBSqlError, nil)
		return
	}

	var metrics []string
	for _, row := range results {
		if row[0] != nil {
			metrics = append(metrics, row[0].(string))
		}
	}

	protocol.WriteQueryResp(w, protocol.CodeOK, metrics)
}
