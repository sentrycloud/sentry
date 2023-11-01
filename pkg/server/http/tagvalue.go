package http

import (
	"database/sql/driver"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"io"
	"net/http"
)

func queryTagValues(w http.ResponseWriter, r *http.Request) {
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

	conn, err := connPool.GetConn()
	if err != nil {
		newlog.Error("queryTagValues: get conn pool failed: %v", err)
		resp.Code = CodeGetConnPoolError
		resp.Msg = CodeMsg[CodeGetConnPoolError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	defer connPool.PutConn(conn)

	rows, err := conn.Query(sql)
	if err != nil {
		newlog.Error("queryTagValues: query taos failed: %v", err)
		resp.Code = CodeExecSqlError
		resp.Msg = CodeMsg[CodeExecSqlError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	defer rows.Close()

	var tagValues []string
	for {
		values := make([]driver.Value, 1)
		if rows.Next(values) == io.EOF {
			break
		}

		tagValues = append(tagValues, values[0].(string))
	}

	resp.Code = CodeOK
	resp.Msg = CodeMsg[CodeOK]
	resp.Data = tagValues
	writeQueryResp(w, http.StatusOK, &resp)
}
