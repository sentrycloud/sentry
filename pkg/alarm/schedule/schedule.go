package schedule

import (
	"github.com/RussellLuo/timingwheel"
	"time"
)

const WheelSize = 256

var tw *timingwheel.TimingWheel

type scheduler struct {
	intervals time.Duration
}

func (s *scheduler) Next(prev time.Time) time.Time {
	next := prev.Add(s.intervals)
	return next
}

func Start() {
	tw = timingwheel.NewTimingWheel(time.Second, WheelSize)
	tw.Start()
}

func Stop() {
	tw.Stop()
}

func Once(delay time.Duration, f func()) *timingwheel.Timer {
	t := tw.AfterFunc(delay, f)
	return t
}

func Repeat(intervals time.Duration, f func()) *timingwheel.Timer {
	s := &scheduler{intervals: intervals}
	t := tw.ScheduleFunc(s, f)
	return t
}
