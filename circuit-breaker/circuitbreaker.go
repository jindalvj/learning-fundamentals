package circuit_breaker

import (
	"errors"
	"sync"
	"time"
)

type State string

const (
	Closed   State = "closed"
	Open     State = "open"
	HalfOpen State = "half-open"
)

type Policy int

const (
	ConsecutiveFails Policy = iota
	TotalFails
)

var ErrCircuitOpen = errors.New("circuit open")

type Config struct {
	Threshold     int
	Timeout       time.Duration
	Policy        Policy
	OnStateChange func(from, to State)
}

type CircuitBreaker struct {
	mu      sync.Mutex
	state   State
	config  Config
	failure int

	openCh chan struct{}
}

func New(config Config) *CircuitBreaker {
	ch := &CircuitBreaker{
		state:  Closed,
		config: config,
		openCh: make(chan struct{}, 1),
	}

	go ch.openWatcher()
	return ch
}

func (ch *CircuitBreaker) openWatcher() {
	for range ch.openCh {
		time.Sleep(ch.config.Timeout)
		ch.mu.Lock()
		ch.changeState(Open, HalfOpen)
		ch.failure = 0
		ch.mu.Unlock()
	}
}

func (ch *CircuitBreaker) changeState(from, to State) {
	if from == to {
		return
	}
	ch.state = to
	if ch.config.OnStateChange != nil {
		ch.config.OnStateChange(from, to)
	}
}

func (ch *CircuitBreaker) Execute(fn func() error) error {
	ch.mu.Lock()

	switch ch.state {
	case Open:
		ch.mu.Unlock()
		return ErrCircuitOpen
	}

	ch.mu.Unlock()

	err := fn()

	ch.mu.Lock()
	defer ch.mu.Unlock()

	if err != nil {
		ch.failure++

		if ch.shouldTrip() {
			ch.trip()
		}
		return err
	}

	if ch.config.Policy == ConsecutiveFails {
		ch.failure = 0
	}
	ch.changeState(ch.state, Closed)
	return nil
}

func (ch *CircuitBreaker) shouldTrip() bool {
	return ch.failure >= ch.config.Threshold
}

func (ch *CircuitBreaker) trip() {
	ch.changeState(ch.state, Open)
	select {
	case ch.openCh <- struct{}{}:
	default:
	}
}

func (ch *CircuitBreaker) State() State {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	return ch.state
}
