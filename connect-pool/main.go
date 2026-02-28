package main

import (
	"connpool/connpool"
	"context"
	"fmt"
	"time"
)

func main() {
	cfg := connpool.DefaultConfig("postgres://user:pass@localhost/mydb?sslmode=disable")
	cfg.MinConns = 2
	cfg.MaxConns = 5

	pool, err := connpool.New(cfg)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	// Acquire a connection (use it, then release via defer)
	db, release, err := pool.Acquire(context.Background())
	if err != nil {
		panic(err)
	}
	defer release() // ALWAYS defer this immediately after Acquire

	var now time.Time
	if err := db.QueryRow("SELECT NOW()").Scan(&now); err != nil {
		panic(err)
	}
	fmt.Println("DB time:", now)
}
