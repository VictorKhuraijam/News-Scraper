package models

import "time"

type Article struct {
    ID        int       `json:"id"`
    SourceID  int       `json:"source_id"`
    Title     string    `json:"title"`
    URL       string    `json:"url"`
    Summary   string    `json:"summary"`
    ScrapedAt time.Time `json:"scraped_at"`
    CreatedAt time.Time `json:"created_at"`
}

type Source struct {
    ID              int       `json:"id"`
    Name            string    `json:"name"`
    URL             string    `json:"url"`
    SelectorTitle   string    `json:"selector_title"`
    SelectorLink    string    `json:"selector_link"`
    SelectorSummary string    `json:"selector_summary"`
    Active          bool      `json:"active"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
