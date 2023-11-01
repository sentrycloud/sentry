package http

import (
	"database/sql/driver"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func queryTopn(w http.ResponseWriter, r *http.Request) {
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

	conn, err := connPool.GetConn()
	if err != nil {
		newlog.Error("queryTopn: get conn pool failed: %v", err)
		resp.Code = CodeGetConnPoolError
		resp.Msg = CodeMsg[CodeGetConnPoolError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	defer connPool.PutConn(conn)

	sql := buildTopnQuerySql(&req)
	rows, err := conn.Query(sql)
	if err != nil {
		newlog.Error("queryTopn: query taos failed: %v", err)
		resp.Code = CodeExecSqlError
		resp.Msg = CodeMsg[CodeExecSqlError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	var topNDataList []protocol.TopNData
	for {
		values := make([]driver.Value, 2) // field, type, length, note
		if rows.Next(values) != nil {
			break
		}

		data := protocol.TopNData{
			Metric: req.Metric,
			Tags:   req.Tags,
			Value:  values[1].(float64),
		}
		data.Tags[req.Field] = values[0].(string)
		topNDataList = append(topNDataList, data)
	}

	_ = rows.Close()

	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = topNDataList
	writeQueryResp(w, http.StatusOK, &resp)
}
