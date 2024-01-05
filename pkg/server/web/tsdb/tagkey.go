package tsdb

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

// QueryTagKeys query all tags of a metric
func QueryTagKeys(w http.ResponseWriter, r *http.Request) {
	var m protocol.MetricReq
	err := protocol.DecodeRequest(r, &m)
	if err != nil {
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	sql := fmt.Sprintf("desc `%s`", m.Metric)
	results, err := QueryTSDB(sql, 4)
	if err != nil {
		protocol.WriteQueryResp(w, protocol.CodeExecTSDBSqlError, nil)
		return
	}

	var tags []string
	for _, row := range results {
		note := row[3].(string)
		if note == "TAG" {
			tags = append(tags, row[0].(string))
		}
	}

	protocol.WriteQueryResp(w, protocol.CodeOK, tags)
}
