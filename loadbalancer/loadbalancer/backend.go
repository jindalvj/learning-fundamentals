package loadbalancer

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	alive        int32
	proxy        *httputil.ReverseProxy
	mu           sync.Mutex
	requestCount int64
}

func NewBackend(rawURL string) (*Backend, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		log.Println(e.Error())
		http.Error(writer, e.Error(), http.StatusBadRequest)
	}

	b := &Backend{URL: u, proxy: proxy}
	b.SetAlive(true)
	return b, nil
}

func (b *Backend) SetAlive(alive bool) {
	if alive {
		atomic.StoreInt32(&b.alive, 1)
	} else {
		atomic.StoreInt32(&b.alive, 0)
	}
}

func (b *Backend) Alive() bool {
	return atomic.LoadInt32(&b.alive) == 1
}

func (b *Backend) RequestCount() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.requestCount
}

func (b *Backend) IncrRequestCount() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.requestCount++
	return b.requestCount
}
