package database

import (
    "context"
    "database/sql"
    "news-scraper/internal/models"
)

// Repository provides database operations
// It abstracts SQL queries and provides a clean interface
type Repository struct {
    db *sql.DB
}

// NewRepository creates a new repository
func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

// GetActiveSources retrieves all active news sources
// Used by scraper to know which sites to scrape
func (r *Repository) GetActiveSources(ctx context.Context) ([]models.Source, error) {
    query := `SELECT id, name, url, selector_title, selector_link, selector_summary, default_category, active, created_at, updated_at FROM sources WHERE active = TRUE ORDER BY name`

    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var sources []models.Source
    for rows.Next() {
        var s models.Source
        err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.SelectorTitle,
            &s.SelectorLink, &s.SelectorSummary, &s.DefaultCategory, &s.Active, &s.CreatedAt, &s.UpdatedAt)
        if err != nil {
            return nil, err
        }
        sources = append(sources, s)
    }
    return sources, rows.Err()
}

// SaveArticle saves an article to the database
// Uses ON DUPLICATE KEY UPDATE to avoid duplicate entries
// If article URL already exists, it updates title and summary
func (r *Repository) SaveArticle(ctx context.Context, article *models.Article) error {
    query := `INSERT INTO articles (source_id, title, url, summary, category, scraped_at)
              VALUES (?, ?, ?, ?, ?, NOW())
              ON DUPLICATE KEY UPDATE title=VALUES(title), summary=VALUES(summary), category = VALUES(category), scraped_at = NOW()`

    _, err := r.db.ExecContext(ctx, query,
        article.SourceID, article.Title, article.URL, article.Summary, article.Category)

    return err
}

// GetRecentArticles retrieves the most recent articles
// Ordered by scraped_at descending (newest first)
func (r *Repository) GetRecentArticles(ctx context.Context, limit int) ([]models.Article, error) {
    query := `SELECT id, source_id, source_name, title, url, summary, category, scraped_at, created_at
              FROM articles ORDER BY scraped_at DESC LIMIT ?`

    rows, err := r.db.QueryContext(ctx, query, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var articles []models.Article
    for rows.Next() {
        var a models.Article
        err := rows.Scan(&a.ID, &a.SourceID, &a.SourceName, &a.Title, &a.URL, &a.Summary, &a.Category, &a.ScrapedAt, &a.CreatedAt)
        if err != nil {
            return nil, err
        }
        articles = append(articles, a)
    }
    return articles, rows.Err()
}

//Get articles by category
func(r *Repository) GetArticlesByCategory(ctx context.Context, category string, limit int ) ([]models.Article, error) {
    query := `SELECT id, source_id, source_name, title, url, summary, category, scraped_at, created_at FROM articles where category = ? ORDER BY scraped_at DESC LIMIT 5 ?`

    rows, err := r.db.QueryContext(ctx, query, category, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var articles []models.Article
    for rows.Next() {
        var a models.Article
        err := rows.Scan(
            &a.ID, &a.SourceID, &a.SourceName, &a.Title, &a.URL, &a.Summary, &a.Category, &a.ScrapedAt, &a.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        articles = append(articles, a)
    }
    return articles, rows.Err()
}

//Get all available categories
func (r *Repository) GetCategories(ctx context.Context) ([]string, error) {
    query := `SELECT DISTINCT category FROM articles WHERE category IS NOT NULL ORDER BY category`

    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var categories []string
    for rows.Next() {
        var cat string
        if err := rows.Scan(&cat); err != nil {
            return nil, err
        }
        categories = append(categories, cat)
    }

    return categories, rows.Err()
}

// GetArticlesBySource retrieves articles from a specific source
func (r *Repository) GetArticlesBySource(ctx context.Context, sourceID int) ([]models.Article, error) {
    query := `SELECT id, source_id, source_name, title, url, summary, category, scraped_at, created_at
              FROM articles WHERE source_id = ? ORDER BY scraped_at DESC LIMIT 50`

    rows, err := r.db.QueryContext(ctx, query, sourceID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var articles []models.Article
    for rows.Next() {
        var a models.Article
        err := rows.Scan(&a.ID, &a.SourceID, &a.SourceName, &a.Title, &a.URL, &a.Summary, &a.Category, &a.ScrapedAt, &a.CreatedAt)
        if err != nil {
            return nil, err
        }
        articles = append(articles, a)
    }
    return articles, rows.Err()
}

// GetSourceByID retrieves a single source by ID
func (r *Repository) GetSourceByID(ctx context.Context, id int) (*models.Source, error) {
    query := `
        SELECT id, name, url, selector_title, selector_link, selector_summary, dafault_category, active, created_at, updated_at
        FROM sources
        WHERE id = ?
    `

    var s models.Source
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &s.ID, &s.Name, &s.URL,
        &s.SelectorTitle, &s.SelectorLink, &s.SelectorSummary, &s.DefaultCategory, &s.Active, &s.CreatedAt, &s.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }

    return &s, nil
}
