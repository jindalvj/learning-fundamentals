package circuit_breaker

import (
	"errors"
	"testing"
	"time"
)

var errService = errors.New("service error")

func fail(cb *CircuitBreaker) error {
	return cb.Execute(func() error { return errService })
}

func succeed(cb *CircuitBreaker) error {
	return cb.Execute(func() error { return nil })
}

// --- circuit breaker tests (same as step 3) ---

func TestTripsAfterThreshold(t *testing.T) {
	cb := New(Config{Threshold: 3, Timeout: time.Second, Policy: ConsecutiveFails})
	fail(cb)
	fail(cb)
	fail(cb)
	if cb.State() != Open {
		t.Fatalf("expected Open, got %s", cb.State())
	}
}

func TestHalfOpenSuccessCloses(t *testing.T) {
	cb := New(Config{Threshold: 3, Timeout: 10 * time.Millisecond, Policy: ConsecutiveFails})
	fail(cb)
	fail(cb)
	fail(cb)
	time.Sleep(time.Second)
	succeed(cb)
	if cb.State() != Closed {
		t.Fatalf("expected Closed, got %s", cb.State())
	}
}
