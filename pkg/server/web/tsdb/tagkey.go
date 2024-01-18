package tsdb

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"github.com/sentrycloud/sentry/pkg/server/monitor"
	"net/http"
	"time"
)

func internalQueryTagKeys(metric string) ([]string, int) {
	sql := fmt.Sprintf("desc `%s`", metric)
	results, err := QueryTSDB(sql, 4)
	if err != nil {
		return nil, protocol.CodeExecTSDBSqlError
	}

	var tags []string
	for _, row := range results {
		note := row[3].(string)
		if note == "TAG" {
			tags = append(tags, row[0].(string))
		}
	}
	return tags, protocol.CodeOK
}

// QueryTagKeys query all tags of a metric
func QueryTagKeys(w http.ResponseWriter, r *http.Request) {
	defer monitor.AddMonitorStats(time.Now(), "tagKey")

	var m protocol.MetricReq
	err := protocol.DecodeRequest(r, &m)
	if err != nil {
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	tags, code := internalQueryTagKeys(m.Metric)
	protocol.WriteQueryResp(w, code, tags)
}
