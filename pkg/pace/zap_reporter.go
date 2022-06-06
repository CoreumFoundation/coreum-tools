package pace

import (
	"strconv"
	"time"

	"go.uber.org/zap"
)

// ZapReporter reports using the provided zap logger and stops reporting when flow of events is stoped.
func ZapReporter(log *zap.Logger) ReporterFunc {
	var previous int
	var stalled time.Time

	return func(label string, timeframe time.Duration, value int) {
		switch {
		case value == 0 && previous == 0:
			return // don't report anything
		case value == 0 && previous != 0:
			dur := timeframe
			if !stalled.IsZero() {
				dur = time.Since(stalled)
				n := dur / timeframe
				if dur-n*timeframe < 10*time.Millisecond {
					dur = n * timeframe
				}
			} else {
				stalled = time.Now().Add(-dur)
			}
			log.Sugar().Infof("%s: stalled for %v", label, dur)
			return
		default:
			previous = value
			stalled = time.Time{}
		}

		floatFmt := func(f float64) string {
			return strconv.FormatFloat(f, 'f', 3, 64)
		}

		intFmt := func(n int) string {
			return strconv.FormatInt(int64(n), 10)
		}

		switch timeframe {
		case time.Second:
			log.Sugar().Infof("%s: %s/s in %v", label, intFmt(value), timeframe)
		case time.Minute:
			log.Sugar().Infof("%s: %s/m in %v", label, intFmt(value), timeframe)
		case time.Hour:
			log.Sugar().Infof("%s: %s/h in %v", label, intFmt(value), timeframe)
		case 24 * time.Hour:
			log.Sugar().Infof("%s: %s/day in %v", label, intFmt(value), timeframe)
		default:
			log.Sugar().Infof("%s %s in %v (pace: %s/s)", intFmt(value), label,
				timeframe, floatFmt(float64(value)/(float64(timeframe)/float64(time.Second))))
		}
	}
}
