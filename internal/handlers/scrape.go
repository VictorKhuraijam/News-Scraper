package handlers

import (
	"context"
	"log"
	"news-scraper/internal/scraper"
	"news-scraper/web/templates"

	"github.com/gofiber/fiber/v2"
)

type ScrapeHandler struct {
    scraper *scraper.Scraper
}

func NewScrapeHandler(scraper *scraper.Scraper) *ScrapeHandler {
    return &ScrapeHandler{scraper: scraper}
}

// TriggerScrape starts scraping in background and returns immediately
// IMPORTANT: Uses context.Background() not c.Context()
// Why? Because:
// 1. c.Context() is cancelled when HTTP response is sent
// 2. Scraping continues in background goroutine after response
// 3. If we used c.Context(), scraping would be cancelled immediately
func (h *ScrapeHandler) TriggerScrape(c *fiber.Ctx) error {
    // Start scraping in background goroutine
    go func() {
        // if err := h.scraper.ScrapeAll(c.Context()); err != nil {
        //     // Log error but don't block response
        // }

        // Create new context that won't be cancelled when HTTP response completes
        ctx := context.Background()// Don't use c.Context() in goroutine
        if err := h.scraper.ScrapeAll(ctx); err != nil {
            log.Printf("Scraping failed: %v", err)
        } else {
            log.Println("Background scraping completed successfully")
        }
    }()

    // return c.JSON(fiber.Map{
    //     "message": "Scraping started",
    // })
    c.Set("Content-Type", "text/html")
    return templates.SuccessMessage("Scraping started in background! Check articles page in a few moments.").Render(c.Context(), c.Response().BodyWriter())


}
