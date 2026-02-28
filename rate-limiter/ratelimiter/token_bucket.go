package ratelimiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	tb.tokens += elapsed * tb.refillRate

	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

func (tb *TokenBucket) AllowN(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	tb.tokens += elapsed * tb.refillRate

	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now

	if tb.tokens >= float64(n) {
		tb.tokens -= float64(n)
		return true
	}
	return false
}
