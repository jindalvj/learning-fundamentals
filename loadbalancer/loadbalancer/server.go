package loadbalancer

import (
	"encoding/json"
	"net/http"
	"time"
)

type Server struct {
	LB *LoadBalancer
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := s.LB.NextBackend()

	if backend == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "No backend available",
		})
		return
	}

	_ = backend.IncrRequestCount()

	w.Header().Set("X-Served-by", backend.URL.Host)
	backend.proxy.ServeHTTP(w, r)
}

func (s *Server) StatusHandler(w http.ResponseWriter, r *http.Request) {
	type backendInfo struct {
		URL          string `json:"url"`
		Alive        bool   `json:"alive"`
		RequestCount int64  `json:"request_count"`
	}

	infos := make([]backendInfo, len(s.LB.backends))
	for i, b := range s.LB.backends {
		infos[i] = backendInfo{
			URL:          b.URL.Host,
			Alive:        b.Alive(),
			RequestCount: b.RequestCount(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"backends":  infos,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
