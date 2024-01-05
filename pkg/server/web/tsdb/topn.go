package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func QueryTopn(w http.ResponseWriter, r *http.Request) {
	var req protocol.TopNRequest
	err := protocol.DecodeRequest(r, &req)
	if err != nil {
		newlog.Error("queryTopn: decode query request failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	code := transferTopnRequest(&req)
	if code != protocol.CodeOK {
		newlog.Error("queryTopn: transferTopnRequest failed: %s", protocol.CodeMsg[code])
		protocol.WriteQueryResp(w, code, nil)
		return
	}

	sql := buildTopnQuerySql(&req)
	results, err := QueryTSDB(sql, 2)
	if err != nil {
		protocol.WriteQueryResp(w, protocol.CodeExecTSDBSqlError, nil)
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

	protocol.WriteQueryResp(w, protocol.CodeOK, topNDataList)
}
