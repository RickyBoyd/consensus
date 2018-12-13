package raft

import "time"

type AgentTimer interface {
	start(duration time.Duration, f func())
	stop()
	restart(duration time.Duration)
	time() <-chan time.Time
	setTimeoutHandler(f func())
}

type RealTimer struct {
	timer *time.Timer
	f     func()
}

func (realTimer *RealTimer) start(duration time.Duration, f func()) {
	realTimer.timer = time.AfterFunc(duration, f)
}

func (realTimer *RealTimer) stop() {
	if !realTimer.timer.Stop() {
		<-realTimer.timer.C
	}
}

func (realTimer *RealTimer) restart(duration time.Duration) {
	realTimer.stop()
	realTimer.timer.Reset(duration)
}

func (realTimer *RealTimer) time() <-chan time.Time {
	return realTimer.timer.C
}

func (realTimer *RealTimer) setTimeoutHandler(f func()) {
	realTimer.f = f
}
