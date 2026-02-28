package ratelimiter

import (
	"sync"
	"time"
)

type SlidingWindowCounter struct {
	prevCounter    int
	currentCounter int
	limit          int
	windowSize     time.Duration
	currentStart   time.Time
	mu             sync.Mutex
}

func NewSlidingWindowCounter(limit int, windowSize time.Duration) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		windowSize:     windowSize,
		prevCounter:    0,
		currentCounter: 0,
		limit:          limit,
		currentStart:   time.Now(),
	}
}

func (c *SlidingWindowCounter) Allow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.currentStart)

	if elapsed > c.windowSize {
		c.prevCounter = c.currentCounter
		c.currentCounter = 0
		c.currentStart = now
	}

	prevWeight := 1 - (elapsed / c.windowSize)
	weightedCount := float64(c.prevCounter)*float64(prevWeight) + float64(c.currentCounter)

	if weightedCount < float64(c.limit) {
		c.currentCounter++
		return true
	}

	return false
}

func (c *SlidingWindowCounter) GetCounters() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.prevCounter, c.currentCounter
}
