package scraper

import (
    "fmt"
    "net/url"
    "strings"

    "github.com/PuerkitoBio/goquery"
)

func ParseArticles(doc *goquery.Document, baseURL, titleSelector, linkSelector, summarySelector string) ([]ScrapedArticle, error) {
    var articles []ScrapedArticle

    doc.Find(titleSelector).Each(func(i int, s *goquery.Selection) {
        var article ScrapedArticle

        // Extract title
        article.Title = strings.TrimSpace(s.Text())
        if article.Title == "" {
            return
        }

        // Extract link
        link, exists := s.Attr("href")
        if !exists {
            // Try finding link in parent or within the selection
            link, exists = s.Find("a").Attr("href")
            if !exists {
                return
            }
        }

        // Resolve relative URLs
        parsedURL, err := url.Parse(link)
        if err != nil {
            return
        }

        base, err := url.Parse(baseURL)
        if err != nil {
            return
        }

        article.URL = base.ResolveReference(parsedURL).String()

        // Extract summary if selector provided
        if summarySelector != "" {
            summary := s.Closest("article, div").Find(summarySelector).First().Text()
            article.Summary = strings.TrimSpace(summary)
        }

        articles = append(articles, article)
    })

    return articles, nil
}
