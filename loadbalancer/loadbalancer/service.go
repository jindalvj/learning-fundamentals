package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func Init() {
	backendURLs := []string{
		"http://localhost:8081",
		"http://localhost:8082",
	}

	backends := make([]*Backend, 0, len(backendURLs))

	for _, url := range backendURLs {
		b, err := NewBackend(url)
		if err != nil {
			log.Fatal(err)
		}

		backends = append(backends, b)
	}

	lb := NewLoadBalancer(backends)

	hc := NewHealthChecker(lb, 10*time.Second)
	hc.Start()

	srv := &Server{LB: lb}

	mux := http.NewServeMux()
	mux.Handle("/", srv)
	mux.HandleFunc("/status", srv.StatusHandler)

	lbPort := 8080

	if err := http.ListenAndServe(fmt.Sprintf(":%d", lbPort), mux); err != nil {
		log.Fatal(err)
	}
}
