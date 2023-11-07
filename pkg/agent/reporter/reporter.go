package reporter

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net"
	"strings"
	"time"
)

const (
	MetricsChanSize     = 1000
	MaxMetricSize       = 8192
	SendMetricBatchSize = 20
)

type Reporter struct {
	serverAddr string
	conn       net.Conn
	localIP    string
	metricList []protocol.MetricValue

	metricsChan chan []protocol.MetricValue
	ticker      *time.Ticker
}

func (r *Reporter) Start(serverAddr string) {
	r.serverAddr = serverAddr

	for {
		// block until connect to sentry server
		r.tryConnect()
		if r.conn != nil {
			r.getLocalIP()
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}

	r.metricsChan = make(chan []protocol.MetricValue, MetricsChanSize)
	r.ticker = time.NewTicker(1 * time.Second)

	go r.listenChanEvents()
}

func (r *Reporter) Report(metrics []protocol.MetricValue) {
	r.metricsChan <- metrics
}

func (r *Reporter) tryConnect() {
	if r.conn == nil {
		conn, err := net.Dial("tcp", r.serverAddr)
		if err != nil {
			newlog.Error("connect to %s failed: %v", r.serverAddr, err)
		} else {
			newlog.Info("connect to %s success", r.serverAddr)
			r.conn = conn
		}
	}
}

func (r *Reporter) getLocalIP() {
	localAddr := r.conn.LocalAddr().String() // ipv4 format: "192.0.2.1:25", ipv6 format: "[2001:db8::1]:80"
	index := strings.LastIndex(localAddr, ":")
	if index != -1 {
		if localAddr[0] != '[' {
			r.localIP = localAddr[0:index]
		} else {
			r.localIP = localAddr[1 : index-1]
		}
	} else {
		newlog.Error("localAddress=%s is not valid", localAddr)
	}

	newlog.Info("LocalIP=%s", r.localIP)
}

func (r *Reporter) listenChanEvents() {
	for {
		select {
		case metrics := <-r.metricsChan:
			r.processMetrics(metrics)
			r.sendMetrics(false)
		case <-r.ticker.C:
			r.sendMetrics(true)
		}
	}
}

func (r *Reporter) processMetrics(metrics []protocol.MetricValue) {
	for _, metric := range metrics {
		if !metric.IsValid() {
			continue
		}

		_, exist := metric.Tags["ip"]
		if !exist {
			metric.Tags["ip"] = r.localIP // if tags not contain ip, agent add local ip to tags
		}

		if len(r.metricList) > MaxMetricSize {
			newlog.Error("discard metric data, cause the cache list is full")
			continue
		}

		r.metricList = append(r.metricList, metric)
	}
}

func (r *Reporter) sendMetrics(fromTicker bool) {
	// if the connection is not valid, only ticker can trigger reconnect, so the reconnect will not too frequently
	if r.conn == nil && fromTicker {
		r.tryConnect()
	}

	if r.conn == nil || len(r.metricList) == 0 {
		return
	}

	// try to send data in batch mode or in every tick, to improve payload send efficiency
	// TODO: if the metricList is too long, split the list and send data
	if fromTicker || len(r.metricList) >= SendMetricBatchSize {
		data, err := protocol.SerializeMetricValues(r.metricList)
		if err != nil {
			newlog.Error("SerializeMetricValues failed: %v", err)
			r.metricList = nil
			return
		}

		n, err := r.conn.Write(data)
		if err != nil || n < len(data) {
			newlog.Error("send metric data failed: %v", err)
			r.conn.Close()
			r.conn = nil
		} else {
			r.metricList = nil
		}
	}
}
