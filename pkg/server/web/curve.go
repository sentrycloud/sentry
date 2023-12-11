package web

import (
	"database/sql/driver"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"io"
	"net/http"
)

func queryCurves(w http.ResponseWriter, r *http.Request) {
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

	conn, err := connPool.GetConn()
	if err != nil {
		newlog.Error("queryCurves: get conn pool failed: %v", err)
		resp.Code = CodeGetConnPoolError
		resp.Msg = CodeMsg[CodeGetConnPoolError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	defer connPool.PutConn(conn)

	rows, err := conn.Query(sql)
	if err != nil {
		newlog.Error("queryCurves: query taos failed: %v", err)
		resp.Code = CodeExecSqlError
		resp.Msg = CodeMsg[CodeExecSqlError]
		writeQueryResp(w, http.StatusBadRequest, &resp)
		return
	}

	defer rows.Close()
	for {
		values := make([]driver.Value, len(starKeys))
		if rows.Next(values) == io.EOF {
			break
		}

		tags := make(map[string]string)
		for idx, key := range starKeys {
			tags[key] = values[idx].(string)
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
