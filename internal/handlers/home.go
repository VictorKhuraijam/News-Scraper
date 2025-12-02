package handlers

import (
	"news-scraper/internal/database"
	"news-scraper/web/templates"

	"github.com/gofiber/fiber/v2"
)

type HomeHandler struct {
	repo *database.Repository
}

func NewHomeHandler(repo *database.Repository) *HomeHandler {
	return &HomeHandler{repo: repo}
}

func (h *HomeHandler) Index(c *fiber.Ctx) error {
	// return c.Render("home", fiber.Map{
	//     "Title": "News Scraper",
	// })
	c.Set("Content-Type", "text/html")
	return templates.Home().Render(c.Context(), c.Response().BodyWriter())
}
