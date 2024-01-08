package merge

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/server/config"
	"github.com/sentrycloud/sentry/pkg/server/monitor"
	"github.com/sentrycloud/sentry/pkg/server/taos"
	"strings"
	"time"
)

type Merge struct {
	conf     config.MergeConfig
	connPool *taos.ConnPool

	totalBufferSize int
	mergeChan       chan string
	resendChan      chan string
	sendTicker      *time.Ticker
	mergeBuffer     *strings.Builder
	resendBuffer    []string
}

func CreateMerge(mergeConfig config.MergeConfig, connPool *taos.ConnPool) *Merge {
	var merge = &Merge{}
	merge.conf = mergeConfig
	merge.connPool = connPool
	merge.mergeBuffer = new(strings.Builder)

	merge.totalBufferSize = 0
	merge.mergeChan = make(chan string, merge.conf.ChanSize)
	merge.resendChan = make(chan string, merge.conf.ChanSize)
	merge.sendTicker = time.NewTicker(time.Duration(merge.conf.TickInterval) * time.Second)
	return merge
}

func (m *Merge) CollectMetrics() {
	monitor.MergeChanSizeCollector.Put(float64(len(m.mergeChan)))
	monitor.ResendChanSizeCollector.Put(float64(len(m.resendChan)))
}

func (m *Merge) MergePayload(payload string) {
	m.mergeChan <- payload
}

func (m *Merge) Start() {
	go m.start()
}

func (m *Merge) start() {
	for {
		select {
		case payload := <-m.mergeChan:
			m.appendPayload(payload)
			m.trySendPayload(false)
		case payload := <-m.resendChan:
			m.resendBuffer = append(m.resendBuffer, payload)
		case <-m.sendTicker.C:
			m.trySendPayload(true)
		}
	}
}

func (m *Merge) appendPayload(payload string) {
	m.totalBufferSize += len(payload)
	if m.mergeBuffer.Len() == 0 {
		m.mergeBuffer.WriteString("[")
	} else {
		m.mergeBuffer.WriteString(",")
	}
	m.mergeBuffer.WriteString(payload)
}

func (m *Merge) trySendPayload(fromTick bool) {
	// no metrics to send to taos server
	if len(m.resendBuffer) == 0 && m.mergeBuffer.Len() == 0 {
		return
	}

	// not accumulated enough metrics
	if !fromTick && len(m.resendBuffer) == 0 && m.mergeBuffer.Len() < m.conf.PayloadBatchSize {
		return
	}

	// resend failed metrics first
	if len(m.resendBuffer) > 0 {
		for _, payload := range m.resendBuffer {
			go m.sendPayload(payload)
		}

		m.resendBuffer = nil
	}

	// send new metrics in batch mode
	if m.mergeBuffer.Len() >= m.conf.PayloadBatchSize || fromTick {
		m.mergeBuffer.WriteString("]") // enclose metric list with [ ]
		go m.sendPayload(m.mergeBuffer.String())
		m.mergeBuffer = new(strings.Builder)

		m.sendTicker.Reset(time.Duration(m.conf.TickInterval) * time.Second)
	}
}

func (m *Merge) sendPayload(payload string) {
	err := m.connPool.SchemalessWrite(payload)
	if err != nil {
		newlog.Error("send payload failed: %v", err)
		m.resendChan <- payload
	}
}
