package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func QueryCurves(w http.ResponseWriter, r *http.Request) {
	var req protocol.MetricReq
	err := protocol.DecodeRequest(r, &req)
	if err != nil {
		newlog.Error("queryCurves: decode query request failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	starTags, noStarTags, err := splitTags(req.Tags)
	if err != nil {
		newlog.Error("queryCurves: splitTags failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeSplitTagsError, nil)
		return
	}

	var curveList []map[string]string
	if len(starTags) == 0 {
		newlog.Info("queryCurves: no star tags ")
		curveList = append(curveList, noStarTags)
		protocol.WriteQueryResp(w, protocol.CodeOK, curveList)
		return
	}

	sql, starKeys := buildCurvesRequest(req.Metric, noStarTags, starTags)
	results, err := QueryTSDB(sql, len(starKeys))
	if err != nil {
		protocol.WriteQueryResp(w, protocol.CodeExecTSDBSqlError, nil)
		return
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

	protocol.WriteQueryResp(w, protocol.CodeOK, curveList)
}
