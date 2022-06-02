package pace

import (
	"sync"
	"sync/atomic"
	"time"
)

type paceAtomic struct {
	mux *sync.RWMutex

	value    int64
	label    string
	paused   bool
	interval time.Duration
	lastTick time.Time
	repFn    ReporterFunc
	t        *time.Timer
}

func (p *paceAtomic) Step(f float64) {
	panic("Step is not implemented in atomic, use StepN")
}

func (p *paceAtomic) StepN(n int) {
	atomic.AddInt64(&p.value, int64(n))
}

func (p *paceAtomic) resetValue() {
	atomic.StoreInt64(&p.value, 0)
}

func (p *paceAtomic) Pause() {
	p.t.Stop()

	p.mux.Lock()
	defer p.mux.Unlock()
	p.report(nil)

	p.paused = true
	p.resetValue()
	p.lastTick = time.Now()
}

func (p *paceAtomic) Resume(interval time.Duration) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.report(nil)

	p.paused = false
	p.resetValue()
	p.lastTick = time.Now()
	if interval > 0 {
		// override the interval if provided
		p.interval = interval
	}

	p.t.Reset(p.interval)
}

func (p *paceAtomic) Report(reporter ReporterFunc) {
	p.t.Stop()
	p.mux.Lock()
	defer p.mux.Unlock()
	p.report(reporter)

	p.resetValue()
	p.lastTick = time.Now()
	if !p.paused {
		p.t.Reset(p.interval)
	}
}

func (p *paceAtomic) report(reporter ReporterFunc) {
	if reporter == nil {
		reporter = p.repFn
	}

	timeframe := time.Since(p.lastTick)
	if abs(timeframe-p.interval) < 10*time.Millisecond {
		timeframe = p.interval
	}

	reporter(p.label, timeframe, float64(atomic.LoadInt64(&p.value)))
}

// New creates a new pace meter (atomic impl) with provided label and reporting function.
// All ticks (or steps) are aggregated in timeframes specified using interval.
//
// paceAtomic is an implementation that uses atomic primitives
// to manage the counter. Offers 3x performance improvement (6ns vs 18ns) per step,
// at expense of not supporting float steps.
func NewAtomic(label string, interval time.Duration, repFn ReporterFunc) Pace {
	if repFn == nil {
		repFn = DefaultReporter()
	}

	p := &paceAtomic{
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

				p.resetValue()
				p.lastTick = time.Now()
				p.t.Reset(interval)
			}()
		}
	}()

	return p
}
