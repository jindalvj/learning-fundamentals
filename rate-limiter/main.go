package main

import (
	"fmt"
	"rate-limiter/ratelimiter"
	"time"
)

func main() {
	fmt.Println("=== Rate Limiter Demonstrations ===\n")

	// 1. Token Bucket Demo
	fmt.Println("1. TOKEN BUCKET (Capacity: 5, Refill: 2 tokens/sec)")
	fmt.Println("   Allows bursts up to capacity, then refills at constant rate")
	tokenBucket := ratelimiter.NewTokenBucket(5, 2)

	// Quick burst of 5 requests (should all succeed)
	fmt.Println("   Burst of 5 requests:")
	for i := 0; i < 5; i++ {
		fmt.Printf("   Request %d: %v\n", i+1, tokenBucket.Allow())
	}

	// 6th request should fail (bucket empty)
	fmt.Printf("   Request 6 (immediate): %v\n", tokenBucket.Allow())

	// Wait for refill
	fmt.Println("   Waiting 1 second for refill...")
	time.Sleep(1 * time.Second)
	fmt.Printf("   Request 7 (after 1s): %v (2 tokens refilled)\n", tokenBucket.Allow())

	fmt.Println()

	// 2. Leaky Bucket Demo
	fmt.Println("2. LEAKY BUCKET (Capacity: 5, Leak: 2 requests/sec)")
	fmt.Println("   Smooths traffic by processing at constant rate")
	leakyBucket := ratelimiter.NewLeakyBucket(5, 2)

	fmt.Println("   Adding 5 requests to queue:")
	for i := 0; i < 5; i++ {
		allowed := leakyBucket.Allow()
		fmt.Printf("   Request %d: %v (Queue size: %d)\n", i+1, allowed, leakyBucket.GetQueueSize())
	}

	// 6th request should fail (queue full)
	fmt.Printf("   Request 6: %v (Queue full)\n", leakyBucket.Allow())

	// Wait for leakage
	fmt.Println("   Waiting 1 second for leak...")
	time.Sleep(1 * time.Second)
	fmt.Printf("   Request 7 (after 1s): %v (2 requests leaked)\n", leakyBucket.Allow())

	fmt.Println()

	// 3. Fixed Window Counter Demo
	fmt.Println("3. FIXED WINDOW COUNTER (Limit: 3 per 5 seconds)")
	fmt.Println("   Simple counter that resets at window boundary")
	fixedWindow := ratelimiter.NewFixedWindowCounter(3, 5*time.Second)

	fmt.Println("   First window:")
	for i := 0; i < 4; i++ {
		allowed := fixedWindow.Allow()
		counter, remaining := fixedWindow.GetCounter()
		fmt.Printf("   Request %d: %v (Count: %d, Window resets in: %v)\n",
			i+1, allowed, counter, remaining.Round(time.Second))
	}

	fmt.Println("   Waiting for window to reset...")
	time.Sleep(5 * time.Second)
	fmt.Printf("   Request 5 (new window): %v\n", fixedWindow.Allow())

	fmt.Println()

	// 4. Sliding Window Log Demo
	fmt.Println("4. SLIDING WINDOW LOG (Limit: 3 per 3 seconds)")
	fmt.Println("   Maintains precise log of all requests")
	slidingLog := ratelimiter.NewSlidingWindowLog(3, 3*time.Second)

	fmt.Println("   Adding 3 requests:")
	for i := 0; i < 3; i++ {
		allowed := slidingLog.Allow()
		fmt.Printf("   Request %d at T+0s: %v (Log size: %d)\n",
			i+1, allowed, slidingLog.GetLogSize())
	}

	// 4th should fail
	fmt.Printf("   Request 4 at T+0s: %v (Limit reached)\n", slidingLog.Allow())

	// Wait 2 seconds (still within window)
	fmt.Println("   Waiting 2 seconds...")
	time.Sleep(2 * time.Second)
	fmt.Printf("   Request 5 at T+2s: %v (Old requests still in window)\n", slidingLog.Allow())

	// Wait another 2 seconds (first requests now expired)
	fmt.Println("   Waiting 2 more seconds...")
	time.Sleep(2 * time.Second)
	fmt.Printf("   Request 6 at T+4s: %v (Old requests expired)\n", slidingLog.Allow())

	fmt.Println()

	// 5. Sliding Window Counter Demo
	fmt.Println("5. SLIDING WINDOW COUNTER (Limit: 5 per 5 seconds)")
	fmt.Println("   Weighted average of previous and current window")
	slidingCounter := ratelimiter.NewSlidingWindowCounter(5, 5*time.Second)

	fmt.Println("   Adding 5 requests at T+0s:")
	for i := 0; i < 5; i++ {
		allowed := slidingCounter.Allow()
		prev, curr := slidingCounter.GetCounters()
		fmt.Printf("   Request %d: %v (Prev: %d, Curr: %d)\n", i+1, allowed, prev, curr)
	}

	// 6th should fail
	fmt.Printf("   Request 6 at T+0s: %v (Limit reached)\n", slidingCounter.Allow())

	// Wait 2.5 seconds (halfway through window)
	fmt.Println("   Waiting 2.5 seconds (50% into window)...")
	time.Sleep(2500 * time.Millisecond)
	prev, curr := slidingCounter.GetCounters()
	fmt.Printf("   Request 7 at T+2.5s: %v (Prev: %d * 0.5 + Curr: %d)\n",
		slidingCounter.Allow(), prev, curr)

	fmt.Println("\n=== Comparison Summary ===")
	fmt.Println("Token Bucket:      Best for APIs needing burst capability")
	fmt.Println("Leaky Bucket:      Best for smooth, constant processing rate")
	fmt.Println("Fixed Window:      Simplest, but has boundary issue")
	fmt.Println("Sliding Log:       Most accurate, but memory intensive")
	fmt.Println("Sliding Counter:   Best balance of accuracy and efficiency")
}
