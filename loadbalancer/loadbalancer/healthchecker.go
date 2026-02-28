package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type HealthChecker struct {
	lb       *LoadBalancer
	interval time.Duration
	client   *http.Client
}

func NewHealthChecker(lb *LoadBalancer, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		lb:       lb,
		interval: interval,
		client:   &http.Client{Timeout: time.Second * 2},
	}
}

func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	go func() {
		for range ticker.C {
			hc.checkAll()
		}
	}()
}

func (hc *HealthChecker) checkAll() {
	for _, b := range hc.lb.backends {
		go hc.check(b)
	}
}

func (hc *HealthChecker) check(backend *Backend) {
	healthURL := fmt.Sprintf("%s/health", backend.URL.String())
	resp, err := hc.client.Get(healthURL)

	if err != nil || resp.StatusCode != http.StatusOK {
		if backend.Alive() {
			backend.SetAlive(false)
			return
		}
	}

	if !backend.Alive() {
		log.Printf("backing up %s", backend.URL.String())
	}

	//backend.SetAlive(true)

}
