package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func internalQueryCurves(req *protocol.MetricReq) ([]map[string]string, int) {
	starTags, noStarTags, err := splitTags(req.Tags)
	if err != nil {
		newlog.Error("queryCurves: splitTags failed: %v", err)
		return nil, protocol.CodeSplitTagsError
	}

	var curveList []map[string]string
	if len(starTags) == 0 {
		curveList = append(curveList, noStarTags)
		return curveList, protocol.CodeOK
	}

	sql, starKeys := buildCurvesRequest(req.Metric, noStarTags, starTags)
	results, err := QueryTSDB(sql, len(starKeys))
	if err != nil {
		return nil, protocol.CodeExecTSDBSqlError
	}

	for _, row := range results {
		tags := make(map[string]string)
		for idx, key := range starKeys {
			tags[key] = row[idx].(string)
		}

		for k, v := range noStarTags {
			tags[k] = v
		}

		curveList = append(curveList, tags)
	}

	return curveList, protocol.CodeOK
}

func QueryCurves(w http.ResponseWriter, r *http.Request) {
	var req protocol.MetricReq
	err := protocol.DecodeRequest(r, &req)
	if err != nil {
		newlog.Error("queryCurves: decode query request failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	curveList, code := internalQueryCurves(&req)
	protocol.WriteQueryResp(w, code, curveList)
}
