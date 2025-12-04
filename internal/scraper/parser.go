package scraper

import (
	// "fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseArticles extracts article information from HTML using CSS selectors
// How it works:
// 1. Find all elements matching titleSelector
// 2. For each element, extract title text and link href
// 3. Resolve relative URLs to absolute URLs
// 4. Optionally extract summary text
func ParseArticles(doc *goquery.Document, baseURL, titleSelector, linkSelector, summarySelector string) ([]ScrapedArticle, error) {
	var articles []ScrapedArticle

	// Find all elements matching the title selector
	doc.Find(titleSelector).Each(func(i int, s *goquery.Selection) {
		var article ScrapedArticle

		// Extract title
		article.Title = strings.TrimSpace(s.Text())
		if article.Title == "" {
			return
		}

		// Extract link - try multiple strategies
		link, exists := s.Attr("href")
		if !exists {
			// If the selector isn't a link itself, try finding a link inside
           	link, exists = s.Find("a").Attr("href")
			if !exists {
				return  // Skip if no link found
			}
		}

		// Resolve relative URLs to absolute URLs
        // Example: "/news/article" becomes "https://example.com/news/article"
		parsedURL, err := url.Parse(link)
		if err != nil {
			return  // Skip if URL is invalid
		}

		base, err := url.Parse(baseURL)
		if err != nil {
			return
		}

        // ResolveReference handles both relative and absolute URLs
		article.URL = base.ResolveReference(parsedURL).String()

		// Extract summary if selector provided
		if summarySelector != "" {
			// Look for summary in parent container or nearby elements
			summary := s.Closest("article, div").Find(summarySelector).First().Text()
			article.Summary = strings.TrimSpace(summary)
		}

		articles = append(articles, article)
	})

	return articles, nil
}
