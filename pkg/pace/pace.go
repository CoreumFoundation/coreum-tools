// Package pace provides a threadsafe counter for measuring ticks in the specified timeframe.
package pace

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Pace is a an interface to register ticks, force reporting and pause/resume the meter.
type Pace interface {
	// Step increments the counter of pace.
	Step(n int)
	// Stop shutdowns reporting, emits a final report for the time passed since previous report.
	Stop()
}

// ReporterFunc defines a function used to report current pace.
type ReporterFunc func(label string, timeframe time.Duration, value int)

type paceImpl struct {
	mux *sync.RWMutex

	value    int64
	label    string
	interval time.Duration
	lastTick time.Time
	repFn    ReporterFunc
	cancelFn func()
	timer    *time.Timer
}

func (p *paceImpl) Step(n int) {
	atomic.AddInt64(&p.value, int64(n))
}

func (p *paceImpl) resetValue() {
	atomic.StoreInt64(&p.value, 0)
}

func (p *paceImpl) Stop() {
	p.cancelFn()
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
		timer:    time.NewTimer(interval),
	}

	paceCtx, cancelFn := context.WithCancel(context.Background())
	p.cancelFn = cancelFn

	go p.reportingLoop(paceCtx)

	return p
}

func (p *paceImpl) reportingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			p.timer.Stop()

			func() {
				p.mux.RLock()
				defer p.mux.RUnlock()
				p.report()
			}()

			// exits the loop
			return
		case <-p.timer.C:
			func() {
				p.mux.Lock()
				defer p.mux.Unlock()
				p.report()

				p.resetValue()
				p.lastTick = time.Now()
				p.timer.Reset(p.interval)
			}()
		}
	}

}

func abs(v time.Duration) time.Duration {
	if v < 0 {
		return -v
	}

	return v
}
