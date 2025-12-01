package database

import (
    "context"
    "database/sql"
    "news-scraper/internal/models"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) GetActiveSources(ctx context.Context) ([]models.Source, error) {
    query := `SELECT id, name, url, selector_title, selector_link, selector_summary,
              active, created_at, updated_at FROM sources WHERE active = TRUE`

    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var sources []models.Source
    for rows.Next() {
        var s models.Source
        err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.SelectorTitle,
            &s.SelectorLink, &s.SelectorSummary, &s.Active, &s.CreatedAt, &s.UpdatedAt)
        if err != nil {
            return nil, err
        }
        sources = append(sources, s)
    }
    return sources, rows.Err()
}

func (r *Repository) SaveArticle(ctx context.Context, article *models.Article) error {
    query := `INSERT INTO articles (source_id, title, url, summary)
              VALUES (?, ?, ?, ?)
              ON DUPLICATE KEY UPDATE title=VALUES(title), summary=VALUES(summary)`

    _, err := r.db.ExecContext(ctx, query,
        article.SourceID, article.Title, article.URL, article.Summary)
    return err
}

func (r *Repository) GetRecentArticles(ctx context.Context, limit int) ([]models.Article, error) {
    query := `SELECT id, source_id, title, url, summary, scraped_at, created_at
              FROM articles ORDER BY scraped_at DESC LIMIT ?`

    rows, err := r.db.QueryContext(ctx, query, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var articles []models.Article
    for rows.Next() {
        var a models.Article
        err := rows.Scan(&a.ID, &a.SourceID, &a.Title, &a.URL,
            &a.Summary, &a.ScrapedAt, &a.CreatedAt)
        if err != nil {
            return nil, err
        }
        articles = append(articles, a)
    }
    return articles, rows.Err()
}

func (r *Repository) GetArticlesBySource(ctx context.Context, sourceID int) ([]models.Article, error) {
    query := `SELECT id, source_id, title, url, summary, scraped_at, created_at
              FROM articles WHERE source_id = ? ORDER BY scraped_at DESC LIMIT 50`

    rows, err := r.db.QueryContext(ctx, query, sourceID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var articles []models.Article
    for rows.Next() {
        var a models.Article
        err := rows.Scan(&a.ID, &a.SourceID, &a.Title, &a.URL,
            &a.Summary, &a.ScrapedAt, &a.CreatedAt)
        if err != nil {
            return nil, err
        }
        articles = append(articles, a)
    }
    return articles, rows.Err()
}
