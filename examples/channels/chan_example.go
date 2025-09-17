package main

import (
	"fmt"
	"math/rand"
	"time"
)

// waitress simulates a waitress placing orders into the orders channel.
// It handles backpressure by retrying if the channel is full, using linear backoff.
func waitress(orders OrdersIn, customers int) {
	for i := 1; i <= customers; i++ {
		order := fmt.Sprintf("Coffee #%d", i) // Create a new coffee order
		select {
		case orders <- order: // Try to send order to barista
			pause := 50 + time.Duration(rand.Intn(100)) // Random delay (50-150ms) - moderate to show flow
			fmt.Println("📝 Taking order, waiting...")   // Simulate order-taking delay
			time.Sleep(pause * time.Millisecond)
			fmt.Println("🗒️ Placed:", order)
		default: // Channel is full (backpressure), retry with backoff
			pause := 200 + time.Duration(rand.Intn(300)) // Backoff delay (200-500ms) - longer to demonstrate backpressure
			fmt.Println("⏳ Barista is full, waiting...") // Indicate backpressure
			time.Sleep(pause * time.Millisecond)         // Linear backoff delay
			i--                                          // Retry the same order
		}
	}
	close(orders) // Close channel after all orders are placed
}

// barista simulates a barista processing orders from the merged channel.
// It brews coffee with a random delay, handles timeouts, and retries failures.
func barista(merged MergedOut, retries RetriesIn, failure int) {
	for coffee := range merged { // Receive orders from merged channel
		pause := 500 + time.Duration(rand.Intn(500)) // Random brewing time (500-1000ms) - longer to trigger timeouts
		fmt.Printf("🍵 Brewing... %s (estimated %dms)\n", coffee, pause)

		// Single timeout with brewing simulation
		timeout := time.After(900 * time.Millisecond) // Timeout at 800ms - allows some timeouts
		brewing := time.After(pause * time.Millisecond)

		select {
		case <-brewing: // Brewing completed on time
			pause := 100 + time.Duration(rand.Intn(100)) // Delivery delay (100-200ms) - short but noticeable
			fmt.Printf("☕ Prepared %s on time, delivering...\n", coffee)
			time.Sleep(pause * time.Millisecond) // Simulate delivery time
		case <-timeout: // Timeout if brewing takes too long
			pause := 100 + time.Duration(rand.Intn(100))                     // Retry delay (100-200ms) - short but noticeable
			fmt.Printf("⏰ Timeout: %s took too long, retrying...\n", coffee) // Simulate retry delay
			select {
			case retries <- coffee: // Try to send to retries channel
				fmt.Printf("🔄 Retrying %s after timeout...\n", coffee)
				time.Sleep(pause * time.Millisecond)
			default:
				fmt.Printf("🚫 Retry queue full after timeout, dropping %s\n", coffee)
			}
			continue // Skip to next order, can't fail if timed out
		}

		// Simulate random delivery failure (after successful brew)
		if rand.Intn(10) < failure { // failure% chance of failure
			fmt.Printf("❌ Failed to deliver %s, retrying...\n", coffee)
			select {
			case retries <- coffee: // Send failed order to retry channel
				pause := 100 + time.Duration(rand.Intn(100)) // Retry delay (100-200ms) - short but noticeable
				fmt.Printf("🔄 Retrying %s...\n", coffee)     // Simulate retry delay
				time.Sleep(pause * time.Millisecond)
			default: // Retry channel full, drop the order
				fmt.Printf("🚫 Retry queue full, dropping %s\n", coffee)
			}
			continue // Skip to next order
		}
		fmt.Printf("✅ Delivered: %s\n", coffee) // Successful delivery
	}
}

// manager merges orders and retries into a single merged channel for the barista.
// It ensures all orders (new and retried) are processed, even after orders channel closes.
func manager(merged MergedIn, orders OrdersOut, retries RetriesOut) {
	defer close(merged)                                       // Close merged channel when done
	for ordersOpen := true; ordersOpen || len(retries) > 0; { // Loop until orders closed and retries empty
		select {
		case order, ok := <-orders: // Receive from orders channel
			if ok {
				merged <- order // Forward order to merged
			} else {
				ordersOpen = false // Orders channel closed
			}
		case retry := <-retries: // Receive from retries channel
			merged <- retry // Forward retry to merged
		}
	}
}

// Type aliases for type safety: enforce send-only or receive-only channels
// Only-Write (send-only)
type OrdersIn chan<- string
type RetriesIn chan<- string
type MergedIn chan<- string

// Only-Read (receive-only)
type OrdersOut <-chan string
type RetriesOut <-chan string
type MergedOut <-chan string

// main sets up the channels, starts goroutines, and runs the barista.
// Demonstrates buffered channels, backpressure, and timeouts.
func main() {
	// Create buffered channels for backpressure handling
	orders := make(chan string, 3)  // Buffer for new orders
	retries := make(chan string, 8) // Buffer for failed retries
	merged := make(chan string, 4)  // Buffer for merged orders

	// Start goroutines: waitress places orders, manager merges them
	go waitress(OrdersIn(orders), 15) // Place 15 orders
	go manager(MergedIn(merged), OrdersOut(orders), RetriesOut(retries))

	// Run barista synchronously (blocks until merged closes)
	barista(MergedOut(merged), RetriesIn(retries), 2) // 20% failure rate
}
