package main

import (
	"fmt"
	"math/rand"
	"time"
)

// waitress simulates order placement with intelligent backpressure handling.
// Uses exponential backoff with jitter when the orders channel is full,
// preventing thundering herd problems and ensuring fair retry distribution.
func waitress(orders OrdersIn, customers int, tracker *RetryTracker) {
	fmt.Println("☕ Waitress started shift, ready to pick orders")
	baseDelay := 50 * time.Millisecond
	maxDelay := 1 * time.Second
	multiplier := 2.0

	for i := 1; i <= customers; i++ {
		order := fmt.Sprintf("Coffee #%d", i)

		// Skip orders that have already exceeded the global retry limit
		if !tracker.ShouldProcess(order) {
			fmt.Printf("🚫 Order %s already has too many attempts (%d), skipping\n",
				order, tracker.Count(order))
			continue
		}

		select {
		case orders <- order:
			// Simulate variable order-taking time
			pause := time.Duration(rand.Intn(100))
			fmt.Println("📝 Taking order, waiting...")
			time.Sleep(pause * time.Millisecond)
			fmt.Printf("🗒️ Placed: %s ⏱️ %dms\n", order, pause)

		default:
			// Channel full: implement exponential backoff with jitter
			currentAttempts := tracker.Count(order)

			// Calculate exponential backoff: baseDelay * multiplier^attempts, capped at maxDelay
			pause := min(
				baseDelay+time.Duration(
					float64(baseDelay)*power(multiplier, float64(currentAttempts)),
				),
				maxDelay,
			)

			// Add ±20% jitter to prevent synchronized retry storms
			jitter := rand.Float64()*0.4 - 0.2
			pause = time.Duration(float64(pause) * (1 + jitter))

			fmt.Printf(
				"⏳ Attempt %d on %s: Barista is busy! Backing off for %vms...\n",
				currentAttempts+1,
				order,
				pause,
			)

			tracker.Increment(order)
			time.Sleep(pause * time.Millisecond)

			// Abort if retry limit exceeded after increment
			if !tracker.ShouldProcess(order) {
				fmt.Printf("❌ Giving up after %d attempts for %s, dropping order!\n",
					tracker.maxRetry, order)
				continue
			}

			i-- // Retry current order by decrementing loop counter
		}
	}

	fmt.Println("👋 Waitress finished taking all orders")
	close(orders) // Signal completion by closing the orders channel
}
