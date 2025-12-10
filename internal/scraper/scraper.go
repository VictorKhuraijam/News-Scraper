package scraper

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/PuerkitoBio/goquery"
    "news-scraper/internal/database"
    "news-scraper/internal/models"
)

// ScrapedArticle represents an article extracted from a webpage
type ScrapedArticle struct {
    Title   string
    URL     string
    Summary string
}

// Scraper coordinates the scraping process
// It manages workers, rate limiting, and HTTP requests
type Scraper struct {
    repo        *database.Repository  // Database access for saving articles
    rateLimiter *RateLimiter          // Controls request rate
    client      *http.Client          // Reusable HTTP client
    userAgent   string                // User-Agent header value
    workers     int                   // Number of concurrent workers
}

// Config holds scraper configuration
type Config struct {
    Workers     int           // Number of concurrent goroutines
    Timeout     time.Duration // HTTP request timeout
    RateLimit   int           // Maximum requests per second
    UserAgent   string        // User-Agent string for requests
}

// NewScraper creates a new scraper instance
func NewScraper(repo *database.Repository, cfg Config) *Scraper {
    return &Scraper{
        repo:        repo,
        rateLimiter: NewRateLimiter(cfg.RateLimit),
        client: &http.Client{
            Timeout: cfg.Timeout,  // Timeout applies to each individual request
        },
        userAgent: cfg.UserAgent,
        workers:   cfg.Workers,
    }
}

// ScrapeAll scrapes all active sources concurrently using a worker pool
// WORKFLOW:
// 1. Fetch active sources from database
// 2. Create channels for work distribution
// 3. Start worker pool
// 4. Distribute sources to workers via channel
// 5. Wait for all workers to complete
// 6. Collect and return results
func (s *Scraper) ScrapeAll(ctx context.Context) error {
    // STEP 1: Get all active sources from database
    sources, err := s.repo.GetActiveSources(ctx)
    if err != nil {
        return fmt.Errorf("failed to get sources: %w", err)
    }

    log.Printf("Starting scraping for %d sources with %d workers", len(sources), s.workers)

    // STEP 2: Create channels for work distribution
    // Jobs channel: Sources to be scraped
    // Buffered size = number of sources (so we can send all without blocking)
    jobs := make(chan models.Source, len(sources))

    // Results channel: Collects errors from workers
    // Buffered size = number of sources (one result per source)
    results := make(chan error, len(sources))

    // STEP 3: Start worker pool
    // WaitGroup tracks how many workers are still running
    var wg sync.WaitGroup

    for i := 0; i < s.workers; i++ {
        wg.Add(1) // Increment counter for this worker

        // Launch worker goroutine
        go func(workerID int) {
            defer wg.Done() // Decrement counter when worker exits

            // Worker loop: process sources until channel is closed
            for source := range jobs {
                log.Printf("Worker %d: scraping %s", workerID, source.Name)

                // Scrape this source
                if err := s.scrapeSource(ctx, source); err != nil {
                    log.Printf("Worker %d: error scraping %s: %v", workerID, source.Name, err)
                    results <- err // Send error to results channel
                } else {
                    results <- nil // Send nil to indicate success
                }
            }
        }(i)
    }

    // STEP 4: Send all sources to workers
    // Workers will pick them up from the channel
    for _, source := range sources {
        jobs <- source
    }
    close(jobs)  // Signal that no more jobs are coming

     // STEP 5: Wait for all workers to finish
    wg.Wait()

    // STEP 6: Close results channel (safe now that all workers are done)
    close(results)

    // STEP 7: Collect all errors
    var errors []error
    for err := range results {
        if err != nil {
            errors = append(errors, err)
        }
    }

    if len(errors) > 0 {
        log.Printf("Scraping completed with %d errors", len(errors))
    } else {
        log.Println("Scraping completed successfully")
    }

    return nil
}

// scrapeSource scrapes a single news source
// WORKFLOW:
// 1. Wait for rate limiter token
// 2. Create HTTP request
// 3. Execute request
// 4. Parse HTML
// 5. Extract articles
// 6. Save to database
func (s *Scraper) scrapeSource(ctx context.Context, source models.Source) error {
   // STEP 1: Rate limiting - wait for token
    // This ensures we don't exceed the configured requests per second
    if err := s.rateLimiter.Wait(ctx); err != nil {
        return fmt.Errorf("rate limiter error: %w", err)
    }

    // STEP 2: Create HTTP request with context
    // Context allows cancellation and timeout
    req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    // Set User-Agent to identify our bot
    req.Header.Set("User-Agent", s.userAgent)

    // STEP 3: Execute HTTP request
    resp, err := s.client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to fetch page: %w", err)
    }
    defer resp.Body.Close()

    // Check HTTP status code
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

     // STEP 4: Parse HTML into goquery document
    // goquery provides jQuery-like syntax for HTML parsing
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to parse HTML: %w", err)
    }

    // STEP 5: Extract articles using CSS selectors
    articles, err := ParseArticles(doc, source.URL,
        source.SelectorTitle, source.SelectorLink, source.SelectorSummary)
    if err != nil {
        return fmt.Errorf("failed to parse articles: %w", err)
    }

    log.Printf("Found %d articles from %s", len(articles), source.Name)

    // STEP 6: Save articles to database
    for _, article := range articles {
        dbArticle := &models.Article{
            SourceID: source.ID,
            Title:    article.Title,
            URL:      article.URL,
            Summary:  article.Summary,
        }

        // SaveArticle has ON DUPLICATE KEY UPDATE
        // So it won't create duplicates if article already exists
        if err := s.repo.SaveArticle(ctx, dbArticle); err != nil {
            log.Printf("Failed to save article: %v", err)
            // Continue processing other articles even if one fails
        }
    }

    return nil
}
