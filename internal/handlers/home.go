package handlers

import (
    "github.com/gofiber/fiber/v2"
    "news-scraper/internal/database"
)

type HomeHandler struct {
    repo *database.Repository
}

func NewHomeHandler(repo *database.Repository) *HomeHandler {
    return &HomeHandler{repo: repo}
}

func (h *HomeHandler) Index(c *fiber.Ctx) error {
    return c.Render("home", fiber.Map{
        "Title": "News Scraper",
    })
}
