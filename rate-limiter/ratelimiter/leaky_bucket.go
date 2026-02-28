package ratelimiter

import (
	"sync"
	"time"
)

type LeakyBucket struct {
	capacity int
	queue    int
	leakRate int
	lastLeak time.Time
	mu       sync.Mutex
}

func NewLeakyBucket(capacity int, leakRate int) *LeakyBucket {
	return &LeakyBucket{
		capacity: capacity,
		queue:    0,
		leakRate: leakRate,
		lastLeak: time.Now(),
	}
}

func (l *LeakyBucket) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastLeak).Seconds()

	leaked := int(elapsed * float64(l.leakRate))

	l.queue -= leaked

	if l.queue < 0 {
		l.queue = 0
	}

	l.lastLeak = now

	if l.queue < l.capacity {
		l.queue++
		return true
	}

	return false
}

func (l *LeakyBucket) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastLeak).Seconds()

	leaked := int(elapsed*float64(l.leakRate)) * n

	l.queue -= leaked

	if l.queue < 0 {
		l.queue = 0
	}

	l.lastLeak = now

	if l.queue < l.capacity {
		l.queue++
		return true
	}

	return false
}

func (l *LeakyBucket) GetQueueSize() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.queue
}
