// Package pace provides a threadsafe counter for measuring ticks in the specified timeframe.
package pace

import (
	"sync"
	"time"
)

// Pace is a an interface to register ticks, force reporting and pause/resume the meter.
type Pace interface {
	// Step increments the counter of pace.
	Step(f float64)
	// StepN increments the counter of pace, using integer N.
	StepN(n int)
	// Pause stops reporting until resumed, all steps continue to be counted.
	Pause()
	// Resume resumes the reporting, starting a report with info since the last tick.
	// Specify a new interval or 0 if you don't want to override it.
	Resume(interval time.Duration)
	// Report manually triggers a report with time frame less than the defined interval.
	// Specify a custom reporter function just for this one report.
	Report(reporter ReporterFunc)
}

// ReporterFunc defines a function used to report current pace.
type ReporterFunc func(label string, timeframe time.Duration, value float64)

type paceImpl struct {
	mux *sync.RWMutex

	value    float64
	label    string
	paused   bool
	interval time.Duration
	lastTick time.Time
	repFn    ReporterFunc
	t        *time.Timer
}

func (p *paceImpl) Step(f float64) {
	p.mux.Lock()
	p.value += f
	p.mux.Unlock()
}

func (p *paceImpl) StepN(n int) {
	p.mux.Lock()
	p.value += float64(n)
	p.mux.Unlock()
}

func (p *paceImpl) Pause() {
	p.t.Stop()

	p.mux.Lock()
	defer p.mux.Unlock()
	p.report(nil)

	p.paused = true
	p.value = 0
	p.lastTick = time.Now()
}

func (p *paceImpl) Resume(interval time.Duration) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.report(nil)

	p.paused = false
	p.value = 0
	p.lastTick = time.Now()
	if interval > 0 {
		// override the interval if provided
		p.interval = interval
	}
	p.t.Reset(p.interval)
}

func (p *paceImpl) Report(reporter ReporterFunc) {
	p.t.Stop()
	p.mux.Lock()
	defer p.mux.Unlock()
	p.report(reporter)

	p.value = 0
	p.lastTick = time.Now()
	if !p.paused {
		p.t.Reset(p.interval)
	}
}

func (p *paceImpl) report(reporter ReporterFunc) {
	if reporter == nil {
		reporter = p.repFn
	}
	timeframe := time.Since(p.lastTick)
	if abs(timeframe-p.interval) < 10*time.Millisecond {
		timeframe = p.interval
	}
	label := p.label
	value := p.value
	reporter(label, timeframe, value)
}

// New creates a new pace meter with provided label and reporting function.
// All ticks (or steps) are aggregated in timeframes specified using interval.
func New(label string, interval time.Duration, repFn ReporterFunc) Pace {
	if repFn == nil {
		repFn = DefaultReporter()
	}
	p := &paceImpl{
		mux: new(sync.RWMutex),

		label:    label,
		interval: interval,
		repFn:    repFn,
		lastTick: time.Now(),
		t:        time.NewTimer(interval),
	}
	go func() {
		for range p.t.C {
			func() {
				p.mux.Lock()
				defer p.mux.Unlock()
				p.report(nil)

				p.value = 0
				p.lastTick = time.Now()
				p.t.Reset(interval)
			}()
		}
	}()
	return p
}

func abs(v time.Duration) time.Duration {
	if v < 0 {
		return -v
	}
	return v
}
