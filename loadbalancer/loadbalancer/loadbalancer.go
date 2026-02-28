package loadbalancer

import "sync/atomic"

type LoadBalancer struct {
	backends []*Backend
	current  uint64
}

func NewLoadBalancer(backends []*Backend) *LoadBalancer {
	return &LoadBalancer{backends: backends}
}

func (lb *LoadBalancer) NextBackend() *Backend {
	n := len(lb.backends)
	if n == 0 {
		return nil
	}

	start := atomic.AddUint64(&lb.current, 1)

	for i := 0; i < n; i++ {
		idx := (start + uint64(i)) % uint64(n)
		b := lb.backends[idx]
		if b.Alive() {
			return b
		}
	}

	return nil
}
