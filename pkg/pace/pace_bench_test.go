package pace

import (
	"testing"
	"time"
)

func BenchmarkStepN(b *testing.B) {
	pace := New("steps", time.Minute, DefaultReporter())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			pace.StepN(1)
		}
	}

	pace.Pause()
}

func BenchmarkAtomicStepN(b *testing.B) {
	pace := NewAtomic("steps", time.Minute, DefaultReporter())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			pace.StepN(1)
		}
	}

	pace.Pause()
}
