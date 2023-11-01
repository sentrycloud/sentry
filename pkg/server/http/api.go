package http

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"github.com/sentrycloud/sentry/pkg/server/collector"
	"github.com/sentrycloud/sentry/pkg/server/config"
	"github.com/sentrycloud/sentry/pkg/server/taos"
	"log"
	"net/http"
	"strconv"
)

// separate connection pool for query and write
var connPool *taos.ConnPool
var serverCollector *collector.Collector

func Start(serverConfig *config.ServerConfig, server *collector.Collector) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", urlNotFound)
	mux.HandleFunc(protocol.PutMetricsUrl, putMetricsHandler)
	mux.HandleFunc(protocol.MetricUrl, queryMetrics)
	mux.HandleFunc(protocol.TagKeyUrl, queryTagKeys)
	mux.HandleFunc(protocol.TagValueUrl, queryTagValues)
	mux.HandleFunc(protocol.CurveUrl, queryCurves)
	mux.HandleFunc(protocol.RangeUrl, queryTimeSeriesDataForRange)
	mux.HandleFunc(protocol.LastUrl, queryTimeSeriesDataForLast)
	mux.HandleFunc(protocol.TopNUrl, queryTopn)

	connPool = taos.CreateConnPool(serverConfig.TaosServer)
	serverCollector = server

	go func() {
		newlog.Info("listen on http port: %d", serverConfig.HttpPort)
		log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(serverConfig.HttpPort), mux))
	}()
}

// all other urls goes here
func urlNotFound(w http.ResponseWriter, r *http.Request) {
	newlog.Error("url not found: %s, from: %s", r.URL.Path, r.RemoteAddr)

	var resp = protocol.QueryResp{
		Code: CodeApiNotFound,
		Msg:  CodeMsg[CodeApiNotFound],
		Data: "",
	}
	writeQueryResp(w, http.StatusBadRequest, &resp)
}

// this api is used for collect metrics from sentry-sdk.
// usually sentry-sdk is connected to sentry-agent to send metrics, but if the machine can't install sentry-agent,
// sentry-sdk can be configured to send metrics directly to sentry-server
func putMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics, err := protocol.CollectHttpMetrics(w, r)
	if err != nil {
		serverCollector.HandleMetrics(metrics)
	}
}
