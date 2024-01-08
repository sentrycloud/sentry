package monitor

import (
	sentrySdk "github.com/sentrycloud/sentry-sdk-go"
	"time"
)

const (
	appName              = "sentry_server"
	httpQpsMetricName    = "sentry_server_http_qps"
	httpRtMetricName     = "sentry_server_http_rt"
	agentCountMetricName = "sentry_server_agent_count"
	dataPointsMetricName = "sentry_server_data_point"
	chanSizeMetricName   = "sentry_server_chan_size"
	collectInterval      = 10
)

var (
	AgentCountCollector     sentrySdk.Collector
	DataPointsCollector     sentrySdk.Collector
	MergeChanSizeCollector  sentrySdk.Collector
	ResendChanSizeCollector sentrySdk.Collector
)

func InitMonitor() {
	AgentCountCollector = sentrySdk.GetCollector(agentCountMetricName, nil, sentrySdk.Sum, collectInterval)
	DataPointsCollector = sentrySdk.GetCollector(dataPointsMetricName, nil, sentrySdk.Sum, collectInterval)

	tags := map[string]string{"chan": "merge"}
	MergeChanSizeCollector = sentrySdk.GetCollector(chanSizeMetricName, tags, sentrySdk.Avg, collectInterval)
	tags["chan"] = "resend"
	ResendChanSizeCollector = sentrySdk.GetCollector(chanSizeMetricName, tags, sentrySdk.Avg, collectInterval)

	sentrySdk.SetReportURL("http://127.0.0.1:51001/server/api/putMetrics") // report to this server
	sentrySdk.StartCollectGC(appName)
}

// AddMonitorStats add rt and qps statistics
func AddMonitorStats(start time.Time, api string) {
	rt := time.Since(start).Milliseconds()

	tags := map[string]string{"appName": appName}
	tags["api"] = api

	httpQpsCollector := sentrySdk.GetCollector(httpQpsMetricName, tags, sentrySdk.Sum, collectInterval)
	httpRtCollector := sentrySdk.GetCollector(httpRtMetricName, tags, sentrySdk.Avg, collectInterval)

	httpQpsCollector.Put(1)
	httpRtCollector.Put(float64(rt))
}
