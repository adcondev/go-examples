package main

import (
	"fmt"
	"sync"
)

// power computes base^exp using iterative multiplication.
// Returns the result as a float64 for exponential backoff calculations.
func power(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// normalizeOrderKey extracts the base order ID from order strings,
// stripping retry suffixes to enable consistent retry counting.
func normalizeOrderKey(order string) string {
	var id, retry int
	if _, err := fmt.Sscanf(order, "Coffee #%d", &id); err == nil {
		return fmt.Sprintf("Coffee #%d", id)
	}
	if _, err := fmt.Sscanf(order, "Coffee #%d (retry %d)", &id, &retry); err == nil {
		return fmt.Sprintf("Coffee #%d", id)
	}
	return order // fallback
}

// RetryTracker provides thread-safe counting of retry attempts per order.
// Prevents infinite retries by enforcing a maximum retry limit.
type RetryTracker struct {
	counts   map[string]int // Maps normalized order keys to retry counts
	maxRetry int            // Maximum allowed retries per order
	mu       sync.Mutex     // Protects concurrent access to counts map
}

// Increment atomically increases the retry count for an order and returns the new count.
func (rt *RetryTracker) Increment(order string) int {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	key := normalizeOrderKey(order)
	rt.counts[key]++
	return rt.counts[key]
}

// Count returns the current retry count for an order without modifying it.
func (rt *RetryTracker) Count(order string) int {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	key := normalizeOrderKey(order)
	return rt.counts[key]
}

// ShouldProcess returns true if the order hasn't exceeded the maximum retry limit.
func (rt *RetryTracker) ShouldProcess(order string) bool {
	count := rt.Count(order)
	return count <= rt.maxRetry
}
