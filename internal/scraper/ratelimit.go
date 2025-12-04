package scraper

import (
    "context"
    "time"
)

// RateLimiter implements a token bucket algorithm to control request rate
// Token Bucket Algorithm:
// - Bucket has capacity = requestsPerSecond
// - Bucket refills at rate = 1 token per (1/requestsPerSecond) seconds
// - Each request consumes 1 token
// - If no tokens available, request waits

type RateLimiter struct {
    tokens chan struct{}  // Channel acts as the token bucket
    rate   time.Duration   // Time between token refills
}


// NewRateLimiter creates a rate limiter that allows requestsPerSecond
// Example: requestsPerSecond=10 means max 10 requests per second
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
    rl := &RateLimiter{
        // Buffered channel = bucket with capacity
        tokens: make(chan struct{}, requestsPerSecond),
        // Calculate interval between tokens
        rate:   time.Second / time.Duration(requestsPerSecond),
    }

    // Fill the bucket initially (all tokens available at start)
    for i := 0; i < requestsPerSecond; i++ {
        rl.tokens <- struct{}{}
    }

   // Start background goroutine to refill tokens
    go rl.refill()

    return rl
}

// refill runs in a goroutine and adds tokens back to the bucket
// This ensures sustained rate limiting over time
func (rl *RateLimiter) refill() {
    ticker := time.NewTicker(rl.rate)
    defer ticker.Stop()

    for range ticker.C {
        select {
        case rl.tokens <- struct{}{}: //Try to add a token
            // Token added successfully
        default:
            // Bucket is full (channel buffer full)
            // This is OK - just skip this refill cycle
        }
    }
}

// Wait blocks until a token is available or context is cancelled
// This is called before each HTTP request
func (rl *RateLimiter) Wait(ctx context.Context) error {
    select {
    case <-rl.tokens: // Consume one token
        return nil
    case <-ctx.Done():  // Context cancelled (timeout or shutdown)

        return ctx.Err()
    }
}
