package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Response struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	ServerID  string `json:"server_id"`
	Port      int    `json:"port"`
	Timestamp string `json:"timestamp"`
}

var (
	port     int
	serverID string
)

func main() {
	log.Printf("Starting on port")
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run response.go [port]")
	}

	var err error

	port, err = strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid port number")
	}

	serverID = fmt.Sprintf("server-%d", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ping", pingHandler)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("[%s] Starting on port %d...\n", serverID, port)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {

	resp := Response{
		Status:    "healthy",
		Message:   "Server is up and running",
		ServerID:  serverID,
		Port:      port,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}

	writeJSON(w, http.StatusOK, resp)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	resp := Response{
		Status:   "ok",
		Message:  "pong",
		ServerID: serverID,
		Port:     port,
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Fatal(err)
	}
}
