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

type ScrapedArticle struct {
    Title   string
    URL     string
    Summary string
}

type Scraper struct {
    repo        *database.Repository
    rateLimiter *RateLimiter
    client      *http.Client
    userAgent   string
    workers     int
}

type Config struct {
    Workers     int
    Timeout     time.Duration
    RateLimit   int
    UserAgent   string
}

func NewScraper(repo *database.Repository, cfg Config) *Scraper {
    return &Scraper{
        repo:        repo,
        rateLimiter: NewRateLimiter(cfg.RateLimit),
        client: &http.Client{
            Timeout: cfg.Timeout,
        },
        userAgent: cfg.UserAgent,
        workers:   cfg.Workers,
    }
}

func (s *Scraper) ScrapeAll(ctx context.Context) error {
    sources, err := s.repo.GetActiveSources(ctx)
    if err != nil {
        return fmt.Errorf("failed to get sources: %w", err)
    }

    log.Printf("Starting scrape for %d sources with %d workers", len(sources), s.workers)

    // Create channels for work distribution
    jobs := make(chan models.Source, len(sources))
    results := make(chan error, len(sources))

    // Start worker pool
    var wg sync.WaitGroup
    for i := 0; i < s.workers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for source := range jobs {
                log.Printf("Worker %d: scraping %s", workerID, source.Name)
                if err := s.scrapeSource(ctx, source); err != nil {
                    log.Printf("Worker %d: error scraping %s: %v", workerID, source.Name, err)
                    results <- err
                } else {
                    results <- nil
                }
            }
        }(i)
    }

    // Send jobs to workers
    for _, source := range sources {
        jobs <- source
    }
    close(jobs)

    // Wait for all workers to finish
    wg.Wait()
    close(results)

    // Check for errors
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

func (s *Scraper) scrapeSource(ctx context.Context, source models.Source) error {
    // Rate limiting
    if err := s.rateLimiter.Wait(ctx); err != nil {
        return fmt.Errorf("rate limiter error: %w", err)
    }

    // Create request with context
    req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("User-Agent", s.userAgent)

    // Execute request
    resp, err := s.client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to fetch page: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    // Parse HTML
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to parse HTML: %w", err)
    }

    // Extract articles
    articles, err := ParseArticles(doc, source.URL,
        source.SelectorTitle, source.SelectorLink, source.SelectorSummary)
    if err != nil {
        return fmt.Errorf("failed to parse articles: %w", err)
    }

    log.Printf("Found %d articles from %s", len(articles), source.Name)

    // Save articles to database
    for _, article := range articles {
        dbArticle := &models.Article{
            SourceID: source.ID,
            Title:    article.Title,
            URL:      article.URL,
            Summary:  article.Summary,
        }

        if err := s.repo.SaveArticle(ctx, dbArticle); err != nil {
            log.Printf("Failed to save article: %v", err)
        }
    }

    return nil
}
