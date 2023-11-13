package collector

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"github.com/sentrycloud/sentry/pkg/server/config"
	"github.com/sentrycloud/sentry/pkg/server/merge"
	"io"
	"net"
	"strings"
	"sync/atomic"
)

const (
	InitialPayloadLength = 4096
	MetricLenLimit       = 192
	TagLenLimit          = 64
)

type Collector struct {
	port         int
	maxConnCount int32

	currentConnCount int32
	listener         net.Listener
	merge            *merge.Merge
}

func (c *Collector) Start(config config.ServerConfig, merger *merge.Merge) {
	c.port = config.TcpPort
	c.maxConnCount = int32(config.MaxConnCount)
	c.merge = merger

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", c.port))
	if err != nil {
		newlog.Fatal("listen on tcp port %d failed: %v", c.port, err)
	}

	newlog.Info("listen on tcp port %d", c.port)
	c.listener = listener
	go c.listen()
}

func (c *Collector) listen() {
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			newlog.Error("accept conn failed: %v", err)
			continue
		}
		if atomic.LoadInt32(&c.currentConnCount) > c.maxConnCount {
			newlog.Error("load capacity protect: exceed maximum concurrent connection: %s", conn.RemoteAddr().String())
			_ = conn.Close()
			continue
		}

		newlog.Info("connect from %v", conn.RemoteAddr())
		atomic.AddInt32(&c.currentConnCount, 1)
		go c.handleConn(conn)
	}
}

func (c *Collector) handleConn(conn net.Conn) {
	defer conn.Close()
	defer atomic.AddInt32(&c.currentConnCount, -1)

	var headerBuf [protocol.PduHeadSize]byte
	var payloadBuf = make([]byte, InitialPayloadLength)
	clientIP := protocol.GetIPFromConnAddr(conn.RemoteAddr().String())

	for {
		_, err := io.ReadFull(conn, headerBuf[:])
		if err != nil {
			newlog.Info("read data from %v failed: %v", conn.RemoteAddr(), err)
			return
		}

		header, err := protocol.DeserializePduHeader(headerBuf[:])
		if err != nil {
			newlog.Error("parse pdu header failed: %v", err)
			return
		}

		if header.PayloadLength > InitialPayloadLength {
			payloadBuf = make([]byte, header.PayloadLength)
		}

		_, err = io.ReadFull(conn, payloadBuf[0:header.PayloadLength])
		if err != nil {
			newlog.Info("read data from connection %v failed: %v", conn.RemoteAddr(), err)
			return
		}

		metrics, err := protocol.UnmarshalPayload(payloadBuf[0:header.PayloadLength])
		if err != nil {
			// no need to close connection when parse payload failed
			newlog.Error("unmarshal payload failed: %v", err)
			continue
		}

		c.HandleMetrics(metrics, clientIP)
	}
}

func (c *Collector) HandleMetrics(metrics []protocol.MetricValue, clientIP string) {
	var filterMetrics []protocol.MetricValue
	for _, metric := range metrics {
		_, exist := metric.Tags["sentryIP"]
		if !exist {
			metric.Tags["sentryIP"] = clientIP // if tags not contain sentryIP, add client ip to tags
		}

		if !metric.IsValid() {
			newlog.Info("invalid metric=%s, tags=%s, value=%f", metric.Metric, metric.Tags, metric.Value)
			continue
		}

		c.transferMetric(&metric)

		filterMetrics = append(filterMetrics, metric)
	}

	if len(filterMetrics) > 0 {
		payload, err := protocol.Json.Marshal(filterMetrics)
		if err != nil {
			newlog.Error("marshal json failed: %v", err)
		} else {
			c.merge.MergePayload(string(payload[1 : len(payload)-1])) // remove [] in the payload
		}
	}
}

func (c *Collector) transferMetric(metric *protocol.MetricValue) {
	if len(metric.Metric) > MetricLenLimit {
		metric.Metric = metric.Metric[:MetricLenLimit]
	}

	for k, v := range metric.Tags {
		// if tag value has both single quote and double quote, it will have Syntax error when writing to TDEngine
		if strings.Contains(v, "'") && strings.Contains(v, "\"") {
			newlog.Error("tag key=%s has invalid value: %s", k, v)
			delete(metric.Tags, k)
		}

		if len(k) > TagLenLimit {
			delete(metric.Tags, k)
			newK := k[:TagLenLimit]
			metric.Tags[newK] = v
		}

		// these are reserved for taos field name, version 2.x or 3.x
		if k == "ts" || k == "value" || k == "_ts" || k == "_value" {
			delete(metric.Tags, k)
			newK := k + "_"
			metric.Tags[newK] = v
		}
	}
}
