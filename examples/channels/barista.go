package main

import (
	"fmt"
	"math/rand"
	"time"
)

// barista processes orders with timeout handling and retry logic.
// Simulates brewing delays, delivery failures, and timeout scenarios,
// demonstrating select-based timeout patterns and error recovery.
func barista(merged MergedOut, retries RetriesIn, failRate int, tracker *RetryTracker) {
	fmt.Println("☕ Barista started shift and ready for orders")
	for coffee := range merged {
		baseOrder := normalizeOrderKey(coffee)

		// Skip orders that have exceeded retry limits
		if !tracker.ShouldProcess(coffee) {
			fmt.Printf("🚫 %s has failed too many times (%d), dropping order!\n",
				baseOrder, tracker.Count(coffee))
			continue
		}

		// Simulate variable brewing time
		pause := time.Duration(rand.Intn(1000))
		fmt.Printf("🍵 Brewing... %s (estimated %dms)\n", coffee, pause)

		// Race brewing time against timeout using select
		timeout := time.After(500 * time.Millisecond)
		brewing := time.After(pause * time.Millisecond)

		select {
		case <-brewing:
			fmt.Printf("☕ Successfully prepared %s in time, delivering now...\n", coffee)

			// Simulate random delivery failure based on failRate
			if rand.Intn(10) < failRate {
				fmt.Printf("❌ Oops! Dropped %s during delivery, need to remake it\n", coffee)
				retryCount := tracker.Increment(coffee)
				retryOrder := fmt.Sprintf("%s (retry %d)", baseOrder, retryCount)

				select {
				case retries <- retryOrder:
					fmt.Printf("🔄 Added %s to retry queue\n", retryOrder)
				default:
					fmt.Printf("🚫 Retry queue full! Had to discard %s\n", coffee)
				}
				continue
			}
			fmt.Printf("✅ Successfully delivered: %s to happy customer!\n", coffee)

		case <-timeout:
			fmt.Printf("⏰ Timeout: %s is taking too long (>500ms), need to restart!\n", coffee)
			retryCount := tracker.Increment(coffee)
			retryOrder := fmt.Sprintf("%s (retry %d)", baseOrder, retryCount)

			select {
			case retries <- retryOrder:
				fmt.Printf("🔄 Added %s to retry queue after timeout\n", retryOrder)
			default:
				fmt.Printf("🚫 Retry queue full after timeout! Had to discard %s\n", coffee)
			}
			continue
		}

	}
	fmt.Println("🏁 Barista finished - no more orders to process")
}
