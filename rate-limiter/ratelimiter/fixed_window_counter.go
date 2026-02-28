package ratelimiter

import (
	"sync"
	"time"
)

type FixedWindowCounter struct {
	limit       int
	counter     int
	windowSize  time.Duration
	windowStart time.Time
	mu          sync.Mutex
}

func NewFixedWindowCounter(limit int, windowSize time.Duration) *FixedWindowCounter {
	return &FixedWindowCounter{
		limit:       limit,
		counter:     0,
		windowSize:  windowSize,
		windowStart: time.Now(),
	}
}

func (c *FixedWindowCounter) Allow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	elapsed := now.Sub(c.windowStart)
	if elapsed >= c.windowSize {
		c.counter = 0
		c.windowStart = now
	}

	if c.counter < c.limit {
		c.counter++
		return true
	}

	return false
}

func (c *FixedWindowCounter) GetCounter() (int, time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := time.Since(c.windowStart)
	remaining := c.windowSize - elapsed

	return c.counter, remaining
}
