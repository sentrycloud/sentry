package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func QueryTagValues(w http.ResponseWriter, r *http.Request) {
	var req protocol.MetricReq
	var resp = protocol.QueryResp{}
	err := protocol.Json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		newlog.Error("queryTagValues: decode query request failed: %v", err)
		resp.Code = CodeJsonDecodeError
		resp.Msg = CodeMsg[CodeJsonDecodeError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	starTags, noStarTags, err := splitTags(req.Tags)
	if err != nil {
		newlog.Error("queryTagValues: splitTags failed: %v", err)
		resp.Code = CodeSplitTagsError
		resp.Msg = CodeMsg[CodeSplitTagsError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	if len(starTags) != 1 {
		newlog.Error("queryTagValues: too many or no star tags ")
		resp.Code = CodeStarKeysError
		resp.Msg = CodeMsg[CodeStarKeysError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	sql, _ := buildCurvesRequest(req.Metric, noStarTags, starTags)
	results, err := QueryTSDB(sql, 1)
	if err != nil {
		resp.Code = CodeExecSqlError
		resp.Msg = CodeMsg[CodeExecSqlError]
		WriteQueryResp(w, http.StatusOK, &resp)
		return
	}

	var tagValues []string
	for _, row := range results {
		if row[0] != nil {
			tagValues = append(tagValues, row[0].(string))
		}
	}

	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = tagValues
	writeQueryResp(w, http.StatusOK, &resp)
}
