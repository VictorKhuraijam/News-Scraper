package handlers

import (
    "github.com/gofiber/fiber/v2"
    "news-scraper/internal/database"
    "news-scraper/web/templates"
    "strconv"
)

type ArticlesHandler struct {
    repo *database.Repository
}

func NewArticlesHandler(repo *database.Repository) *ArticlesHandler {
    return &ArticlesHandler{repo: repo}
}

// GetRecent returns articles as JSON (for API)
func (h *ArticlesHandler) GetRecent(c *fiber.Ctx) error {
    limit := 50
    articles, err := h.repo.GetRecentArticles(c.Context(), limit)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch articles",
        })
    }

    return c.JSON(articles)
}

// GetBySource returns articles from specific source as JSON (for API)
func (h *ArticlesHandler) GetBySource(c *fiber.Ctx) error {
    sourceID, err := strconv.Atoi(c.Params("sourceId"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid source ID",
        })
    }

    articles, err := h.repo.GetArticlesBySource(c.Context(), sourceID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch articles",
        })
    }

    // return c.JSON(articles)
    c.Set("Content-Type", "text/html")
    return templates.Articles(articles).Render(c.Context(), c.Response().BodyWriter())
}


// RenderArticles renders the full articles page with Templ
func (h *ArticlesHandler) RenderArticles(c *fiber.Ctx) error {
    limit := 50
    articles, err := h.repo.GetRecentArticles(c.Context(), limit)
    if err != nil {
        return c.Status(500).SendString("Failed to load articles")
    }

    // return c.Render("articles", fiber.Map{
    //     "Articles": articles,
    // })
    // Render full page with layout
    c.Set("Content-Type", "text/html")
    return templates.Articles(articles).Render(c.Context(), c.Response().BodyWriter())
}

// RenderArticlesList renders just the articles list (for HTMX partial updates)
func (h *ArticlesHandler) RenderArticlesList(c *fiber.Ctx) error {
    limit := 50
    articles, err := h.repo.GetRecentArticles(c.Context(), limit)
    if err != nil {
        return c.Status(500).SendString("Failed to load articles")
    }

    // Render only the article list component (no layout)
    c.Set("Content-Type", "text/html")
    return templates.ArticlesList(articles).Render(c.Context(), c.Response().BodyWriter())
}
