package ratelimiter

import (
	"sync"
	"time"
)

type SlidingWindowLog struct {
	limit      int
	windowSize time.Duration
	log        []time.Time
	mu         sync.Mutex
}

func NewSlidingWindowLog(limit int, windowSize time.Duration) *SlidingWindowLog {
	return &SlidingWindowLog{
		limit:      limit,
		windowSize: windowSize,
		log:        make([]time.Time, 0),
	}
}

func (l *SlidingWindowLog) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.windowSize)

	logs := []time.Time{}

	for _, t := range l.log {
		if t.After(cutoff) {
			logs = append(logs, t)
		}
	}

	l.log = logs

	if len(logs) < l.limit {
		l.log = append(l.log, now)
		return true
	}

	return false
}

func (l *SlidingWindowLog) GetLogSize() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.log)
}
