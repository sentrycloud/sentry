package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func QueryCurves(w http.ResponseWriter, r *http.Request) {
	var req protocol.MetricReq
	var resp = protocol.QueryResp{}
	err := protocol.Json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		newlog.Error("queryCurves: decode query request failed: %v", err)
		resp.Code = CodeJsonDecodeError
		resp.Msg = CodeMsg[CodeJsonDecodeError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	starTags, noStarTags, err := splitTags(req.Tags)
	if err != nil {
		newlog.Error("queryCurves: splitTags failed: %v", err)
		resp.Code = CodeSplitTagsError
		resp.Msg = CodeMsg[CodeSplitTagsError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	var curveList []map[string]string
	if len(starTags) == 0 {
		newlog.Info("queryCurves: no star tags ")
		curveList = append(curveList, noStarTags)

		resp.Code = CodeOK
		resp.Msg = CodeMsg[CodeOK]
		resp.Data = curveList
		writeQueryResp(w, http.StatusOK, &resp)
		return
	}

	sql, starKeys := buildCurvesRequest(req.Metric, noStarTags, starTags)
	results, err := QueryTSDB(sql, len(starKeys))
	if err != nil {
		resp.Code = CodeExecSqlError
		resp.Msg = CodeMsg[CodeExecSqlError]
		WriteQueryResp(w, http.StatusOK, &resp)
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

	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = curveList
	writeQueryResp(w, http.StatusOK, &resp)
}
