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

    return c.JSON(articles)
}

func (h *ArticlesHandler) RenderArticles(c *fiber.Ctx) error {
    limit := 50
    articles, err := h.repo.GetRecentArticles(c.Context(), limit)
    if err != nil {
        return c.Status(500).SendString("Failed to load articles")
    }

    // return c.Render("articles", fiber.Map{
    //     "Articles": articles,
    // })
    return templates.Articles(articles).Render(c.Context(), c.Response().BodyWriter())
}

func (h *ArticlesHandler) RenderArticlesList(c *fiber.Ctx) error {
    limit := 50
    articles, err := h.repo.GetRecentArticles(c.Context(), limit)
    if err != nil {
        return c.Status(500).SendString("Failed to load articles")
    }

    // Render just the articles list component for HTMX updates
    return templates.ArticlesList(articles).Render(c.Context(), c.Response().BodyWriter())
}
