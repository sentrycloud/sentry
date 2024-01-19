package monitor

import (
	"fmt"
	sentrySdk "github.com/sentrycloud/sentry-sdk-go"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"net"
	"time"
)

const (
	defaultLocalIP       = "127.0.0.1"
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

	reportURL := fmt.Sprintf("http://%s:51001/server/api/putMetrics", getLocalIP())
	sentrySdk.SetReportURL(reportURL) // report to self
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

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		newlog.Info("get local ip err: ", err)
		return defaultLocalIP
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return defaultLocalIP
}
