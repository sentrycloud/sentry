package web

import (
	"database/sql/driver"
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"io"
	"net/http"
)

// query all metrics that start with the name in the request
// TODO: all query process have some common code that can be extract by functions or macro, but I think the code is more readable in the current form
func queryMetrics(w http.ResponseWriter, r *http.Request) {
	var m protocol.MetricReq
	var resp = protocol.QueryResp{}
	err := protocol.Json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		newlog.Error("queryMetrics: decode query request failed: %v", err)
		resp.Code = CodeJsonDecodeError
		resp.Msg = CodeMsg[CodeJsonDecodeError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	conn, err := connPool.GetConn()
	if err != nil {
		newlog.Error("queryMetrics: get conn pool failed: %v", err)
		resp.Code = CodeGetConnPoolError
		resp.Msg = CodeMsg[CodeGetConnPoolError]
		writeQueryResp(w, http.StatusInternalServerError, &resp)
		return
	}

	defer connPool.PutConn(conn)

	sql := fmt.Sprintf("show stables like '%%%s%%'", m.Metric)
	rows, err := conn.Query(sql)
	if err != nil {
		newlog.Error("queryMetrics: query taos failed: %v", err)
		resp.Code = CodeExecSqlError
		resp.Msg = CodeMsg[CodeExecSqlError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	defer rows.Close()

	var metrics []string
	for {
		values := make([]driver.Value, 1)
		if rows.Next(values) == io.EOF {
			break
		}

		metrics = append(metrics, values[0].(string))
	}

	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = metrics
	writeQueryResp(w, http.StatusOK, &resp)
}
