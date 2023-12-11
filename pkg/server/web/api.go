package web

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"github.com/sentrycloud/sentry/pkg/server/collector"
	"github.com/sentrycloud/sentry/pkg/server/config"
	"github.com/sentrycloud/sentry/pkg/server/taos"
	"github.com/sentrycloud/sentry/pkg/server/web/mysql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// separate connection pool for query
var connPool *taos.ConnPool
var serverCollector *collector.Collector

// SPAHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type SPAHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Join internally call path.Clean to prevent directory traversal
	path := filepath.Join(h.staticPath, r.URL.Path)

	// check whether a file exists or is a directory at the given path
	fi, err := os.Stat(path)
	if os.IsNotExist(err) || fi.IsDir() {
		// file does not exist or path is a directory, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	}

	if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static file
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func Start(serverConfig *config.ServerConfig, server *collector.Collector) {
	spaHandler := SPAHandler{staticPath: "../sentry-frontend/build", indexPath: "index.html"}

	mux := http.NewServeMux()
	mux.Handle("/", spaHandler)

	mux.HandleFunc(protocol.PutMetricsUrl, putMetricsHandler)
	mux.HandleFunc(protocol.MetricUrl, queryMetrics)
	mux.HandleFunc(protocol.TagKeyUrl, queryTagKeys)
	mux.HandleFunc(protocol.TagValueUrl, queryTagValues)
	mux.HandleFunc(protocol.CurveUrl, queryCurves)
	mux.HandleFunc(protocol.RangeUrl, queryTimeSeriesDataForRange)
	mux.HandleFunc(protocol.TopNUrl, queryTopn)

	mux.HandleFunc(protocol.ContactUrl, mysql.HandleContact)

	connPool = taos.CreateConnPool(serverConfig.TaosServer)
	serverCollector = server

	newlog.Info("listen on http port: %d", serverConfig.HttpPort)
	go func() {
		log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(serverConfig.HttpPort), mux))
	}()
}

// this api is used for collect metrics from sentry-sdk.
// usually sentry-sdk is connected to sentry-agent to send metrics, but if the machine can't install sentry-agent,
// sentry-sdk can be configured to send metrics directly to sentry-server
func putMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics, err := protocol.CollectHttpMetrics(w, r)
	if err == nil {
		remoteIP := protocol.GetIPFromConnAddr(r.RemoteAddr)
		serverCollector.HandleMetrics(metrics, remoteIP)
	}
}
