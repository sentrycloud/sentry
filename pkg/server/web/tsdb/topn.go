package tsdb

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func internalQueryTopN(req *protocol.TopNRequest) ([]protocol.TopNData, int) {
	code := transferTopnRequest(req)
	if code != protocol.CodeOK {
		newlog.Error("queryTopN: transferTopNRequest failed: %s", protocol.CodeMsg[code])
		return nil, code
	}

	sql := buildTopnQuerySql(req)
	results, err := QueryTSDB(sql, 2)
	if err != nil {
		return nil, protocol.CodeExecTSDBSqlError
	}

	var topNDataList []protocol.TopNData
	for _, row := range results {
		data := protocol.TopNData{
			Metric: req.Metric,
			Name:   row[0].(string),
			Tags:   req.Tags,
			Value:  row[1].(float64),
		}
		data.Tags[req.Field] = row[0].(string)
		topNDataList = append(topNDataList, data)
	}

	return topNDataList, protocol.CodeOK
}

func QueryTopN(w http.ResponseWriter, r *http.Request) {
	var req protocol.TopNRequest
	err := protocol.DecodeRequest(r, &req)
	if err != nil {
		newlog.Error("queryTopN: decode query request failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	topNDataList, code := internalQueryTopN(&req)
	protocol.WriteQueryResp(w, code, topNDataList)
}
