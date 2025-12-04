
# HOW EVERYTHING WORKS TOGETHER:

1. RATE LIMITER (ratelimit.go)
   - Token Bucket Algorithm
   - Creates buffered channel with N tokens (N = requests per second)
   - Background goroutine refills 1 token every 1/N seconds
   - Wait() consumes 1 token (blocks if none available)
   - Ensures max N requests per second across all workers

2. PARSER (parser.go)
   - Uses goquery (jQuery-like library for Go)
   - Finds HTML elements using CSS selectors
   - Extracts text content and attributes
   - Resolves relative URLs to absolute URLs
   - Returns structured article data

3. SCRAPER (scraper.go)
   - Worker Pool Pattern:
     * Creates N worker goroutines
     * Distributes sources via buffered channel
     * Each worker: rate limit → HTTP request → parse → save
     * WaitGroup ensures all workers finish before returning

4. REPOSITORY (repository.go)
   - Database abstraction layer
   - Uses prepared statements for SQL injection protection
   - ON DUPLICATE KEY UPDATE prevents duplicate articles
   - Context-aware for cancellation support

5. HANDLERS (handlers/*.go)
   - HTTP request handlers
   - Bridge between web requests and business logic
   - Render Templ templates or return JSON
   - Use context.Background() for background jobs

# EXECUTION FLOW:
User clicks "Scrape Now"
  → TriggerScrape handler
  → Starts goroutine with context.Background()
  → ScrapeAll() orchestrates workers
  → Workers scrape sources concurrently
  → Rate limiter controls request rate
  → Parser extracts articles from HTML
  → Repository saves to database
  → User refreshes articles page to see results

KEY CONCURRENCY PATTERNS:
- Goroutines: Lightweight threads
- Channels: Communication between goroutines
- WaitGroups: Wait for multiple goroutines
- Context: Cancellation and timeouts
- Mutexes: Not needed here (channels handle synchronization)
