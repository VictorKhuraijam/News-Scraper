package scraper

import (
	"context"
	"fmt"
	"log"
	"strings"

	"sync"
	"time"

	// "github.com/PuerkitoBio/goquery"
	"news-scraper/internal/database"
	"news-scraper/internal/models"

	"github.com/gocolly/colly/v2"
)

// Scraper coordinates the scraping process
type Scraper struct {
    repo        *database.Repository  // Database access for saving articles
    userAgent   string                // User-Agent header value
    workers     int                   // Number of concurrent workers
    timeout     time.Duration
    rateLimit   int
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
        userAgent:   cfg.UserAgent,
        workers:     cfg.Workers,
        timeout:     cfg.Timeout,
        rateLimit:   cfg.RateLimit,
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
                if err := s.scrapeSourceWithColly(ctx, source); err != nil {
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

// scrapeSource scrapes a single news source using colly
func (s *Scraper) scrapeSourceWithColly(ctx context.Context, source models.Source) error {
  // Track found articles
    var articles []models.Article
    var mu sync.Mutex // Protect articles slice from concurrent access

    // Create a new Colly collector
    c := colly.NewCollector(
        // Set user agent
        colly.UserAgent(s.userAgent),

        // Visit only the specified domain
        colly.AllowedDomains(extractDomain(source.URL)),

        // Enable async mode for better performance
        colly.Async(false),
    )

    // Configure rate limiting (requests per second)
    c.Limit(&colly.LimitRule{
        DomainGlob:  "*",
        Parallelism: 1,
        Delay:       time.Second / time.Duration(s.rateLimit),
    })

    // Set timeout
    c.SetRequestTimeout(s.timeout)

    // Before making a request
    c.OnRequest(func(r *colly.Request) {
        log.Printf("Visiting %s", r.URL.String())
    })

    // On every HTML element matching the selector
    c.OnHTML(source.SelectorTitle, func(e *colly.HTMLElement) {
        article := models.Article{
            SourceID: source.ID,
        }

        // Extract title
        article.Title = e.Text

        // Extract link - try multiple methods
        if source.SelectorLink != "" && source.SelectorLink != source.SelectorTitle {
            // If link selector is different, find it
            linkElem := e.DOM.Find(source.SelectorLink)
            if href, exists := linkElem.Attr("href"); exists {
                article.URL = e.Request.AbsoluteURL(href)
            }
        } else {
            // Link is in the same element
            if href, exists := e.DOM.Attr("href"); exists {
                article.URL = e.Request.AbsoluteURL(href)
            } else {
                // Try finding <a> tag inside
                if href, exists := e.DOM.Find("a").Attr("href"); exists {
                    article.URL = e.Request.AbsoluteURL(href)
                }
            }
        }

        // Extract summary if selector provided
        if source.SelectorSummary != "" {
            // Look for summary in parent or nearby elements
            summaryElem := e.DOM.Closest("article, div").Find(source.SelectorSummary)
            article.Summary = summaryElem.First().Text()
        }

        // ← NEW: Detect and set category
        article.Category = detectCategory(article.Title, article.Summary, article.URL, source.DefaultCategory)


        // Only save if we have both title and URL
        if article.Title != "" && article.URL != "" {
            mu.Lock()
            articles = append(articles, article)
            mu.Unlock()
        }
    })

    // On error
    c.OnError(func(r *colly.Response, err error) {
        log.Printf("Error scraping %s: %v", r.Request.URL, err)
    })

    // On response
    c.OnResponse(func(r *colly.Response) {
        log.Printf("Response from %s: %d bytes", r.Request.URL, len(r.Body))
    })

    // Visit the URL
    if err := c.Visit(source.URL); err != nil {
        return fmt.Errorf("failed to visit %s: %w", source.URL, err)
    }

    // Wait for all async requests to complete
    c.Wait()

    log.Printf("Found %d articles from %s", len(articles), source.Name)

    // Save articles to database
    for _, article := range articles {

        dbArticle := &models.Article{
            SourceID:   source.ID,
            SourceName: source.Name,
            Title:      article.Title,
            URL:        article.URL,
            Summary:    article.Summary,
            Category:   article.Category,
        }
        if err := s.repo.SaveArticle(ctx, dbArticle); err != nil {
            log.Printf("Failed to save article: %v", err)
        }
    }
    // fmt.Printf("Scraped articles is %v", articles)

    return nil
}

//Helper function for category detection
func detectCategory(title, summary, url string, defaultCategory string) string {
    text := strings.ToLower(title + " " + summary + " " + url)

    // Technology keywords
    techKeywords := []string{"tech", "ai", "software", "app", "startup", "code", "programming",
                             "computer", "gadget", "robot", "crypto", "blockchain"}
    for _, kw := range techKeywords {
        if strings.Contains(text, kw) {
            return "technology"
        }
    }

    // Sports keywords
    sportsKeywords := []string{"sport", "football", "soccer", "basketball", "tennis", "cricket",
                               "olympics", "championship", "match", "player", "team", "goal"}
    for _, kw := range sportsKeywords {
        if strings.Contains(text, kw) {
            return "sports"
        }
    }

    // Politics keywords
    politicsKeywords := []string{"politic", "election", "government", "president", "minister",
                                 "parliament", "vote", "law", "senate", "congress"}
    for _, kw := range politicsKeywords {
        if strings.Contains(text, kw) {
            return "politics"
        }
    }

    // Business keywords
    businessKeywords := []string{"business", "market", "stock", "economy", "trade", "finance",
                                 "bank", "investor", "revenue", "profit"}
    for _, kw := range businessKeywords {
        if strings.Contains(text, kw) {
            return "business"
        }
    }

    // Entertainment keywords
    entertainmentKeywords := []string{"entertainment", "movie", "music", "celebrity", "film",
                                     "actor", "actress", "concert", "album", "show"}
    for _, kw := range entertainmentKeywords {
        if strings.Contains(text, kw) {
            return "entertainment"
        }
    }

    // Health keywords
    healthKeywords := []string{"health", "medical", "doctor", "hospital", "disease", "vaccine",
                               "treatment", "patient", "medicine"}
    for _, kw := range healthKeywords {
        if strings.Contains(text, kw) {
            return "health"
        }
    }

    // Default to source's default category
    return defaultCategory
}

// extractDomain extracts domain from URL for Colly's AllowedDomains
func extractDomain(urlStr string) string {
    // Simple domain extraction
    // For "https://techcrunch.com/news" returns "techcrunch.com"
    if len(urlStr) == 0 {
        return "*"
    }

    // Remove protocol
    start := 0
    if idx := indexOf(urlStr, "://"); idx != -1 {
        start = idx + 3
    }

    // Find end of domain
    end := len(urlStr)
    if idx := indexOfFrom(urlStr, "/", start); idx != -1 {
        end = idx
    }

    domain := urlStr[start:end]

    // Remove port if present
    if idx := indexOf(domain, ":"); idx != -1 {
        domain = domain[:idx]
    }

    return domain
}

func indexOf(s, substr string) int {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return i
        }
    }
    return -1
}

func indexOfFrom(s, substr string, from int) int {
    for i := from; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return i
        }
    }
    return -1
}
