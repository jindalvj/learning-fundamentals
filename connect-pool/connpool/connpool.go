package connpool

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrPoolClosed  = errors.New("pool closed")
	ErrPoolTimeout = errors.New("pool timeout")
	ErrInvalidConn = errors.New("invalid connection")
)

type Config struct {
	DSN              string
	MinConns         int
	MaxConns         int
	AcquireTimeout   time.Duration
	MaxConnLifetime  time.Duration
	HealthCheckEvery time.Duration
}

func DefaultConfig(dsn string) *Config {
	return &Config{
		DSN:              dsn,
		MinConns:         2,
		MaxConns:         10,
		AcquireTimeout:   5 * time.Second,
		MaxConnLifetime:  5 * time.Second,
		HealthCheckEvery: 10 * time.Second,
	}
}

type conn struct {
	db        *sql.DB
	createdAt time.Time
	lastUseAt time.Time
}

func (c *conn) isAlive() bool {
	return c.db.PingContext(context.Background()) == nil
}

func (c *conn) isExpired(maxLifetime time.Duration) bool {
	return time.Since(c.createdAt) > maxLifetime
}

type Pool struct {
	cfg     Config
	mu      sync.Mutex
	idle    []*conn
	active  int
	waiters []waiter
	closed  bool
	done    chan struct{}
}

type waiter struct {
	ch  chan *conn
	ctx context.Context
}

func New(cfg Config) (*Pool, error) {
	if cfg.MaxConns < cfg.MinConns {
		return nil, fmt.Errorf("max connections must be greater or equal to min connections")
	}

	p := &Pool{
		cfg:  cfg,
		done: make(chan struct{}),
	}

	for i := 0; i < cfg.MinConns; i++ {
		c, err := p.newConn()
		if err != nil {
			p.closeAll()
			return nil, err
		}

		p.idle = append(p.idle, c)
		p.active++
	}

	go p.healthCheckLoop()

	return p, nil
}

func (p *Pool) Acquire(ctx context.Context) (*sql.DB, func(), error) {
	ctx, cancel := context.WithTimeout(ctx, p.cfg.AcquireTimeout)
	defer cancel()

	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return nil, nil, ErrPoolClosed
	}

	for len(p.idle) > 0 {
		c := p.idle[len(p.idle)-1]
		p.idle = p.idle[:len(p.idle)-1]

		if c.isExpired(p.cfg.MaxConnLifetime) || !c.isAlive() {
			p.active--
			c.db.Close()
			continue
		}

		p.mu.Unlock()
		release := p.releaseFunc(c)
		return c.db, release, nil
	}

	if p.active < p.cfg.MaxConns {
		p.active++
		p.mu.Unlock()

		c, err := p.newConn()
		if err != nil {
			p.mu.Lock()
			p.active--
			p.mu.Unlock()
			return nil, nil, err
		}

		release := p.releaseFunc(c)
		return c.db, release, nil
	}

	ch := make(chan *conn, 1)
	w := waiter{ch: ch, ctx: ctx}
	p.waiters = append(p.waiters, w)
	p.mu.Unlock()

	select {
	case c := <-ch:
		if c == nil {
			return nil, nil, ErrInvalidConn
		}
		release := p.releaseFunc(c)
		return c.db, release, nil
	case <-ctx.Done():
		p.mu.Lock()
		p.removeWaiter(ch)
		p.mu.Unlock()

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, nil, ErrPoolTimeout
		}
		return nil, nil, ctx.Err()
	}
}

func (p *Pool) releaseFunc(c *conn) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			c.lastUseAt = time.Now()
			p.release(c)
		})
	}
}

func (p *Pool) release(c *conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		p.active--
		c.db.Close()
		return
	}

	for len(p.waiters) > 0 {
		w := p.waiters[0]
		p.waiters = p.waiters[1:]

		select {
		case <-w.ctx.Done():
			continue
		default:
			w.ch <- c
			return
		}
	}

	p.idle = append(p.idle, c)
}

func (p *Pool) healthCheckLoop() {
	ticker := time.NewTicker(p.cfg.HealthCheckEvery)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.healthCheck()
		case <-p.done:
			return
		}
	}
}

func (p *Pool) healthCheck() {
	p.mu.Lock()
	defer p.mu.Unlock()

	alive := p.idle[:0]

	for _, c := range p.idle {
		if c.isExpired(p.cfg.MaxConnLifetime) || !c.isAlive() {
			p.active--
			c.db.Close()
			continue
		}
		alive = append(alive, c)
	}

	p.idle = alive

	for p.active < p.cfg.MinConns {
		c, err := p.newConn()
		if err != nil {
			break
		}

		p.idle = append(p.idle, c)
		p.active++
	}
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true
	close(p.done)

	for _, w := range p.waiters {
		w.ch <- nil
	}
	p.waiters = nil
	p.closeAll()
}

func (p *Pool) closeAll() {
	for _, c := range p.idle {
		c.db.Close()
	}
	p.idle = nil
}

func (p *Pool) removeWaiter(ch chan *conn) {
	for i, w := range p.waiters {
		if w.ch == ch {
			p.waiters = append(p.waiters[:i], p.waiters[i+1:]...)
			return
		}
	}
}

func (p *Pool) newConn() (*conn, error) {
	db, err := sql.Open("postgres", p.cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(p.cfg.MaxConns)
	db.SetMaxIdleConns(p.cfg.MaxConns)

	ctx, cancel := context.WithTimeout(context.Background(), p.cfg.AcquireTimeout)

	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return &conn{db: db, createdAt: time.Now(), lastUseAt: time.Now()}, nil
}
