package scraper

import (
	"context"
	"log"
	"time"

	"news-scraper/internal/models"

	"github.com/gocolly/colly/v2"
)

// ScrapeWithPagination scrapes multiple pages (for sites with pagination)
func (s *Scraper) ScrapeWithPagination(ctx context.Context, source models.Source, maxPages int) error {
    c := colly.NewCollector(
         colly.UserAgent(s.userAgent),
         colly.AllowedDomains(extractDomain(source.URL)),
         colly.MaxDepth(maxPages),
    )

    // Configure rate limiting
    c.Limit(&colly.LimitRule{
        DomainGlob:  "*",
        Parallelism: 1,
        Delay:       s.timeout / time.Duration(s.rateLimit),
    })

    pageCount := 0

    // Find articles on each page
    c.OnHTML(source.SelectorTitle, func(e *colly.HTMLElement) {
        article := models.Article{
            SourceID: source.ID,
            Title:    e.Text,
        }

        if href, exists := e.DOM.Attr("href"); exists {
            article.URL = e.Request.AbsoluteURL(href)
        } else if href, exists := e.DOM.Find("a").Attr("href"); exists {
            article.URL = e.Request.AbsoluteURL(href)
        }

        if article.Title != "" && article.URL != "" {
            if err := s.repo.SaveArticle(ctx, &article); err != nil {
                log.Printf("Failed to save article: %v", err)
            }
        }
    })

    // Find next page link (common patterns)
    c.OnHTML("a[rel='next'], a.next, a.pagination-next, .next-page a", func(e *colly.HTMLElement) {
        if pageCount < maxPages {
            nextPage := e.Attr("href")
            if nextPage != "" {
                pageCount++
                log.Printf("Following to page %d: %s", pageCount+1, nextPage)
                e.Request.Visit(nextPage)
            }
        }
    })

    c.OnError(func(r *colly.Response, err error) {
        log.Printf("Error on %s: %v", r.Request.URL, err)
    })

    return c.Visit(source.URL)
}

// ScrapeWithJavaScript scrapes sites that require JavaScript
// NOTE: Requires chromedp or similar for full JS support
func (s *Scraper) ScrapeWithCache(ctx context.Context, source models.Source, cacheDir string) error {
    c := colly.NewCollector(
        colly.UserAgent(s.userAgent),
        colly.AllowedDomains(extractDomain(source.URL)),
        colly.CacheDir(cacheDir), // Enable caching
    )

    c.Limit(&colly.LimitRule{
        DomainGlob:  "*",
        Parallelism: 1,
        Delay:       time.Second / time.Duration(s.rateLimit),
    })

    c.OnHTML(source.SelectorTitle, func(e *colly.HTMLElement) {
        article := models.Article{
            SourceID: source.ID,
            Title:    e.Text,
        }

        if href, exists := e.DOM.Find("a").Attr("href"); exists {
            article.URL = e.Request.AbsoluteURL(href)
        }

        if article.Title != "" && article.URL != "" {
            s.repo.SaveArticle(ctx, &article)
        }
    })

    return c.Visit(source.URL)
}
