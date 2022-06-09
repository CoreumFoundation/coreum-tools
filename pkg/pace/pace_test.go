package pace

import (
	"context"
	"testing"
	"time"
)

func TestSimple(t *testing.T) {
	var ticks int
	var stopped bool

	testReportingFn := func(_ string, _ time.Duration, value int) {
		ticks += value

		if stopped {
			t.Fatalf("not expected to fire after stop")
			t.FailNow()
		}
	}
	p := New(context.Background(), "items", 100*time.Millisecond, testReportingFn)

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

	p.Stop()

	time.Sleep(100 * time.Millisecond)

	stopped = true

	time.Sleep(200 * time.Millisecond)
}
