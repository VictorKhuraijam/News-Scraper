package handlers

import (
	"fmt"
	"news-scraper/internal/database"
	"news-scraper/web/templates"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ArticlesHandler struct {
    repo *database.Repository
}

func NewArticlesHandler(repo *database.Repository) *ArticlesHandler {
    return &ArticlesHandler{repo: repo}
}

// GetRecent returns articles as JSON (for API)
func (h *ArticlesHandler) GetRecent(c *fiber.Ctx) error {
    limit := 100
    articles, err := h.repo.GetRecentArticles(c.Context(), limit)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch articles",
        })
    }

    // return c.JSON(articles)
    c.Set("Content-Type", "text/html")
    return templates.Articles(articles).Render(c.Context(), c.Response().BodyWriter())
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
    limit := 100
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
    limit := 100
    articles, err := h.repo.GetRecentArticles(c.Context(), limit)
    if err != nil {
        return c.Status(500).SendString("Failed to load articles")
    }

    // Render only the article list component (no layout)
    c.Set("Content-Type", "text/html")
    return templates.ArticlesList(articles).Render(c.Context(), c.Response().BodyWriter())
}


// NEW: Get articles by category
// func (h *ArticlesHandler) GetByCategory(c *fiber.Ctx) error {
//     category := c.Params("category")
//     limit := 50

//     articles, err := h.repo.GetArticlesByCategory(c.Context(), category, limit)
//     if err != nil {
//         return c.Status(500).JSON(fiber.Map{
//             "error": "Failed to fetch articles",
//         })
//     }

//     return c.JSON(articles)
// }

// NEW: Render articles by category
func (h *ArticlesHandler) RenderArticlesByCategory(c *fiber.Ctx) error {
    category := c.Params("category")
    limit := 50

    articles, err := h.repo.GetArticlesByCategory(c.Context(), category, limit)
    if err != nil {
        fmt.Println("Render articles by category error :", err)
        return c.Status(500).SendString("Failed to load articles")
    }

    c.Set("Content-Type", "text/html")
    return templates.ArticlesWithCategory(articles, category).Render(
        c.Context(),
        c.Response().BodyWriter(),
    )
}

// NEW: Get all categories
func (h *ArticlesHandler) GetCategories(c *fiber.Ctx) error {
    categories, err := h.repo.GetCategories(c.Context())
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch categories",
        })
    }

    return c.JSON(fiber.Map{
        "categories": categories,
    })
}
