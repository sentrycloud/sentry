package script

import (
	"encoding/json"
	"github.com/sentrycloud/sentry/pkg/agent/reporter"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"time"
)

type Scheduler struct {
	ticker *time.Ticker
	done   chan struct{}
	script *Script
	report *reporter.Reporter
}

func NewScheduler(s *Script, r *reporter.Reporter) *Scheduler {
	scheduler := &Scheduler{script: s, report: r}
	scheduler.ticker = time.NewTicker(time.Duration(s.interval) * time.Second)
	scheduler.done = make(chan struct{})
	return scheduler
}

func (s *Scheduler) Start() {
	newlog.Info("start scheduler for: %s", s.script.path)
	go func() {
		s.run() // ticker will be triggered after plugin.Interval seconds, so we call it manually here
		for {
			select {
			case <-s.ticker.C:
				s.run()
			case <-s.done:
				s.ticker.Stop()
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	newlog.Info("stop scheduler for: %s", s.script.path)
	close(s.done)
}

func (s *Scheduler) run() {
	out, err := RunCommand(s.script.interval-1, s.script.scriptType, s.script.path)
	if err != nil {
		newlog.Error("run %s failed, out=%s, error=%v", s.script.path, string(out), err)
		return
	}

	if len(out) == 0 {
		return
	}

	var values []protocol.MetricValue
	err = json.Unmarshal(out, &values)
	if err != nil {
		newlog.Error("json unmarshal failed: script=%s, out=%s, error=%v", s.script.path, string(out), err)
		return
	}

	s.report.Report(values)
}
