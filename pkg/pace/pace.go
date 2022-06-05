// Package pace provides a threadsafe counter for measuring ticks in the specified timeframe.
package pace

import (
	"sync"
	"sync/atomic"
	"time"
)

// Pace is a an interface to register ticks, force reporting and pause/resume the meter.
type Pace interface {
	// Step increments the counter of pace.
	Step(n int)
	// Pause stops reporting until resumed, all steps continue to be counted.
	Pause()
}

// ReporterFunc defines a function used to report current pace.
type ReporterFunc func(label string, timeframe time.Duration, value int)

type paceImpl struct {
	mux *sync.RWMutex

	value    int64
	label    string
	paused   bool
	interval time.Duration
	lastTick time.Time
	repFn    ReporterFunc
	ticker   *time.Ticker
}

func (p *paceImpl) Step(n int) {
	atomic.AddInt64(&p.value, int64(n))
}

func (p *paceImpl) resetValue() {
	atomic.StoreInt64(&p.value, 0)
}

func (p *paceImpl) Pause() {
	p.ticker.Stop()

	p.mux.Lock()
	defer p.mux.Unlock()
	p.report()

	p.paused = true
	p.resetValue()
	p.lastTick = time.Now()
}

func (p *paceImpl) report() {
	timeframe := time.Since(p.lastTick)
	if abs(timeframe-p.interval) < 10*time.Millisecond {
		timeframe = p.interval
	}

	p.repFn(p.label, timeframe, int(atomic.LoadInt64(&p.value)))
}

// New creates a new pace meter with provided label and reporting function.
// All ticks (or steps) are aggregated in timeframes specified using interval.
func New(label string, interval time.Duration, repFn ReporterFunc) Pace {
	if repFn == nil {
		panic("nil repFn provided")
	}

	p := &paceImpl{
		mux: new(sync.RWMutex),

		label:    label,
		interval: interval,
		repFn:    repFn,
		lastTick: time.Now(),
		ticker:   time.NewTicker(interval),
	}

	go p.reportingLoop()

	return p
}

func (p *paceImpl) reportingLoop() {
	for range p.ticker.C {
		func() {
			p.mux.Lock()
			defer p.mux.Unlock()
			p.report()

			p.resetValue()
			p.lastTick = time.Now()
		}()
	}
}

func abs(v time.Duration) time.Duration {
	if v < 0 {
		return -v
	}

	return v
}
