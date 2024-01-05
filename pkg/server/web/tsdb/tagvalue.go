package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func QueryTagValues(w http.ResponseWriter, r *http.Request) {
	var req protocol.MetricReq
	err := protocol.DecodeRequest(r, &req)
	if err != nil {
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	starTags, noStarTags, err := splitTags(req.Tags)
	if err != nil {
		newlog.Error("queryTagValues: splitTags failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeSplitTagsError, nil)
		return
	}

	if len(starTags) != 1 {
		newlog.Error("queryTagValues: too many or no star tags ")
		protocol.WriteQueryResp(w, protocol.CodeStarKeysError, nil)
		return
	}

	sql, _ := buildCurvesRequest(req.Metric, noStarTags, starTags)
	results, err := QueryTSDB(sql, 1)
	if err != nil {
		protocol.WriteQueryResp(w, protocol.CodeExecTSDBSqlError, nil)
		return
	}

	var tagValues []string
	for _, row := range results {
		if row[0] != nil {
			tagValues = append(tagValues, row[0].(string))
		}
	}

	protocol.WriteQueryResp(w, protocol.CodeOK, tagValues)
}
