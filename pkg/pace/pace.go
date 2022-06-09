// Package pace provides a threadsafe counter for measuring ticks in the specified timeframe.
package pace

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
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
	lastTickMux *sync.RWMutex
	lastTick    time.Time

	label    string
	interval time.Duration
	value    int64

	reportFn      ReporterFunc
	defaultLogger *zap.Logger
	cancelFn      context.CancelFunc
	timer         *time.Timer
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

	p.reportFn(p.label, timeframe, int(atomic.LoadInt64(&p.value)))
}

// New creates a new pace meter with provided label and optional reporting function.
// All ticks (or steps) are aggregated in timeframes specified using interval.
// If the reporting function was not provided, ZapReporter will be used as default.
func New(ctx context.Context, label string, interval time.Duration, reportFn ...ReporterFunc) Pace {
	p := &paceImpl{
		lastTickMux: new(sync.RWMutex),
		lastTick:    time.Now(),

		label:    label,
		interval: interval,

		timer: time.NewTimer(interval),
	}

	logger, _ := zap.NewProduction()
	p.defaultLogger = logger.With(zap.String("label", p.label))

	if len(reportFn) > 0 {
		p.reportFn = reportFn[0]
	} else {
		p.reportFn = ZapReporter(p.defaultLogger)
	}

	paceCtx, cancelFn := context.WithCancel(ctx)
	p.cancelFn = cancelFn

	go p.reportingLoop(paceCtx)

	return p
}

func (p *paceImpl) reportingLoop(ctx context.Context) {
	defer func() {
		if v := recover(); v != nil {
			if err, ok := v.(error); ok {
				p.defaultLogger.With(zap.Error(err)).Warn("pace reportingLoop panicked")
				return
			}

			p.defaultLogger.With(zap.Error(
				errors.Errorf("error: %v", v),
			)).Warn("pace reportingLoop panicked")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			p.timer.Stop()

			func() {
				p.lastTickMux.RLock()
				defer p.lastTickMux.RUnlock()
				p.report()
			}()

			// exits the loop
			return
		case <-p.timer.C:
			func() {
				p.lastTickMux.Lock()
				defer p.lastTickMux.Unlock()
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
