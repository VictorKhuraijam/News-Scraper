package handlers

import (
    "github.com/gofiber/fiber/v2"
    "news-scraper/internal/scraper"
)

type ScrapeHandler struct {
    scraper *scraper.Scraper
}

func NewScrapeHandler(scraper *scraper.Scraper) *ScrapeHandler {
    return &ScrapeHandler{scraper: scraper}
}

func (h *ScrapeHandler) TriggerScrape(c *fiber.Ctx) error {
    go func() {
        if err := h.scraper.ScrapeAll(c.Context()); err != nil {
            // Log error but don't block response
        }
    }()

    return c.JSON(fiber.Map{
        "message": "Scraping started",
    })
}
