package http

import (
	"database/sql/driver"
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"io"
	"net/http"
)

// query all tags of a metric
func queryTagKeys(w http.ResponseWriter, r *http.Request) {
	var m protocol.MetricReq
	var resp = protocol.QueryResp{}
	err := protocol.Json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		newlog.Error("queryTagKeys: decode query request failed: %v", err)
		resp.Code = CodeJsonDecodeError
		resp.Msg = CodeMsg[CodeJsonDecodeError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	conn, err := connPool.GetConn()
	if err != nil {
		newlog.Error("queryTagKeys: get conn pool failed: %v", err)
		resp.Code = CodeGetConnPoolError
		resp.Msg = CodeMsg[CodeGetConnPoolError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	defer connPool.PutConn(conn)

	sql := fmt.Sprintf("desc `%s`", m.Metric)
	rows, err := conn.Query(sql)
	if err != nil {
		newlog.Error("queryTagKeys: query taos failed: %v", err)
		resp.Code = CodeExecSqlError
		resp.Msg = CodeMsg[CodeExecSqlError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	defer rows.Close()

	var tags []string
	for {
		values := make([]driver.Value, 4) // field, type, length, note
		if rows.Next(values) == io.EOF {
			break
		}

		note := values[3].(string)
		if note == "TAG" {
			tags = append(tags, values[0].(string))
		}
	}

	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = tags
	writeQueryResp(w, http.StatusOK, &resp)
}
