package scraper

import (
    "context"
    "time"
)

type RateLimiter struct {
    tokens chan struct{}
    rate   time.Duration
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
    rl := &RateLimiter{
        tokens: make(chan struct{}, requestsPerSecond),
        rate:   time.Second / time.Duration(requestsPerSecond),
    }

    // Fill the bucket initially
    for i := 0; i < requestsPerSecond; i++ {
        rl.tokens <- struct{}{}
    }

    // Refill tokens at the specified rate
    go rl.refill()

    return rl
}

func (rl *RateLimiter) refill() {
    ticker := time.NewTicker(rl.rate)
    defer ticker.Stop()

    for range ticker.C {
        select {
        case rl.tokens <- struct{}{}:
        default:
            // Bucket is full
        }
    }
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    select {
    case <-rl.tokens:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
