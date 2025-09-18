// Package main demonstrates Go concurrency patterns through a coffee shop simulation.
// Features showcased:
// - Buffered channels for async communication and backpressure handling
// - Select statements for non-blocking operations and timeouts
// - Exponential backoff with jitter for retry logic
// - Type-safe directional channels (send-only vs receive-only)
// - Coordinated goroutine shutdown patterns
package main

import (
	"fmt"
	"sync"
)

// Channel type aliases enforce directional usage at compile-time.
// This prevents accidentally sending to receive-only channels or vice versa,
// making the code more robust and self-documenting.

// Send-only channel types (data flows INTO these channels)
type OrdersIn chan<- string  // Waitress sends orders →
type RetriesIn chan<- string // Barista sends failed orders →
type MergedIn chan<- string  // Manager sends merged orders →
type DeadIn chan<- string    // Manager sends dead letters →

// Receive-only channel types (data flows OUT OF these channels)
type OrdersOut <-chan string  // Manager receives from orders ←
type RetriesOut <-chan string // Manager receives from retries ←
type MergedOut <-chan string  // Barista receives from merged ←

// main orchestrates the coffee shop simulation with proper goroutine coordination.
// Sets up buffered channels, launches worker goroutines, and waits for completion.
func main() {
	tracker := &RetryTracker{
		counts:   make(map[string]int),
		maxRetry: 2, // Allow up to 3 total attempts per order
	}

	var wg sync.WaitGroup

	fmt.Println("🏪 Coffee Shop Simulation Starting")
	fmt.Println("=================================")

	// Create buffered channels with different capacities for realistic flow control
	orders := make(chan string, 10)  // New orders queue
	retries := make(chan string, 10) // Failed order retry queue
	merged := make(chan string, 10)  // Combined work queue for barista
	dead := make(chan string, 10)    // Dead letter queue for failed orders

	fmt.Printf("📊 Channel capacities - Orders: %d, Retries: %d, Merged: %d, Dead: %d\n",
		cap(orders), cap(retries), cap(merged), cap(dead))

	// Launch all goroutines with proper coordination
	wg.Add(4)
	fmt.Println("🚀 Starting all goroutines")

	go func() {
		defer wg.Done()
		waitress(OrdersIn(orders), 20, tracker)
	}()

	go func() {
		defer wg.Done()
		manager(MergedIn(merged), OrdersOut(orders), RetriesOut(retries), DeadIn(dead), tracker)
		close(dead) // Close dead letters when manager completes
	}()

	go func() {
		defer wg.Done()
		barista(MergedOut(merged), RetriesIn(retries), 2, tracker) // 20% failure rate
	}()

	go func() {
		defer wg.Done()
		// Process dead letters (orders that exceeded retry limits)
		for order := range dead {
			fmt.Printf("📮 Dead Letter: %s moved to failed orders log\n", order)
		}
	}()

	// Wait for all goroutines to complete their work
	wg.Wait()

	fmt.Println("=================================")
	fmt.Println("🎉 Coffee Shop Simulation Complete!")
}
