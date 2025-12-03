package handlers

import (
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

func (h *ScrapeHandler) TriggerScrape(c *fiber.Ctx) error {
    go func() {
        if err := h.scraper.ScrapeAll(c.Context()); err != nil {
            // Log error but don't block response
        }
    }()

    // return c.JSON(fiber.Map{
    //     "message": "Scraping started",
    // })
    c.Set("Content-Type", "text/html")
    return templates.SuccessMessage("Scraping started in background!").Render(c.Context(), c.Response().BodyWriter())


}
