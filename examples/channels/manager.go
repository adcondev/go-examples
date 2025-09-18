package main

import (
	"fmt"
	"time"
)

// manager coordinates order flow between new orders and retries.
// Merges multiple input channels into a single output channel while handling
// graceful shutdown and ensuring all work completes before termination.
func manager(merged MergedIn, orders OrdersOut, retries RetriesOut, dead DeadIn, tracker *RetryTracker) {
	fmt.Println("👨‍💼 Manager started coordinating orders and retries")
	defer close(merged)

	// Timeout prevents hanging when waiting for final retries
	noActivityTimer := time.NewTimer(1 * time.Second)
	defer noActivityTimer.Stop()

	// Process until orders close AND retry queue empties
	for ordersOpen := true; ordersOpen || len(retries) > 0; {
		noActivityTimer.Reset(1 * time.Second)
		select {
		case order, ok := <-orders:
			if ok {
				fmt.Printf("👨‍💼 Manager: Forwarding new order: %s\n", order)
				merged <- order
			} else {
				fmt.Println("📢 Manager: No more new orders coming in (orders channel closed)")
				ordersOpen = false
			}

		case retry := <-retries:
			if tracker.ShouldProcess(retry) {
				merged <- retry
				fmt.Printf("👨‍💼 Manager: Handling retry for: %s\n", retry)
			} else {
				// Move failed orders to dead letter queue
				select {
				case dead <- retry:
					fmt.Printf("⚰️ Order %s exceeded retry limit, moved to dead letters\n",
						normalizeOrderKey(retry))
				default:
					fmt.Printf("💀 Dead letter queue full, discarding %s\n",
						normalizeOrderKey(retry))
				}
			}

		case <-noActivityTimer.C:
			// Handle graceful shutdown after orders close
			if !ordersOpen {
				if len(retries) == 0 {
					fmt.Println("👨‍💼 Manager: Orders done, retries empty, shutting down")
					return
				}
			}
		}
	}
	fmt.Println("👨‍💼 Manager finished - all orders and retries processed")
}
