package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func QueryTopn(w http.ResponseWriter, r *http.Request) {
	var req protocol.TopNRequest
	var resp = protocol.QueryResp{}
	err := protocol.Json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		newlog.Error("queryTopn: decode query request failed: %v", err)
		resp.Code = CodeJsonDecodeError
		resp.Msg = CodeMsg[CodeJsonDecodeError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	code := transferTopnRequest(&req)
	if code != CodeOK {
		newlog.Error("queryTopn: transferTopnRequest failed: %s", CodeMsg[code])
		resp.Code = code
		resp.Msg = CodeMsg[code]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	sql := buildTopnQuerySql(&req)
	results, err := QueryTSDB(sql, 2)
	if err != nil {
		resp.Code = CodeExecSqlError
		resp.Msg = CodeMsg[CodeExecSqlError]
		WriteQueryResp(w, http.StatusOK, &resp)
		return
	}

	var topNDataList []protocol.TopNData
	for _, row := range results {
		data := protocol.TopNData{
			Metric: req.Metric,
			Tags:   req.Tags,
			Value:  row[1].(float64),
		}
		data.Tags[req.Field] = row[0].(string)
		topNDataList = append(topNDataList, data)
	}
	
	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = topNDataList
	writeQueryResp(w, http.StatusOK, &resp)
}
