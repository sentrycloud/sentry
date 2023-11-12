package httpcollector

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/agent/reporter"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

var agentReporter *reporter.Reporter

func putMetricsHandler(w http.ResponseWriter, req *http.Request) {
	metrics, err := protocol.CollectHttpMetrics(w, req)
	if err == nil {
		agentReporter.Report(metrics)
	}
}

func Start(report *reporter.Reporter, httpPort int) {
	agentReporter = report

	mux := http.NewServeMux()
	mux.HandleFunc("/agent/api/putMetrics", putMetricsHandler)

	newlog.Info("Listen on http port %d", httpPort)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", httpPort), mux)
		if err != nil {
			newlog.Fatal("Listen on http port %d failed: %v", httpPort, err)
		}
	}()
}
