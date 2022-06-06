package pace

import (
	"testing"
	"time"
)

func TestSimple(t *testing.T) {
	var ticks int

	p := New("items", 100*time.Millisecond, func(_ string, _ time.Duration, value int) {
		ticks += value
	})

	go func() {
		for i := 0; i < 10; i++ {
			p.Step(1)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	if ticks != 0 {
		t.Fatalf("expeted ticks: 0, got: %d", ticks)
		t.FailNow()
	}

	time.Sleep(60 * time.Millisecond)

	if ticks != 10 {
		t.Fatalf("expeted ticks: 10, got: %d", ticks)
		t.FailNow()
	}
}
