package scheduler

import (
	"context"
	"log"

	"news-scraper/internal/database"
	"news-scraper/internal/scraper"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
    cron    *cron.Cron
    scraper *scraper.Scraper
    repo    *database.Repository
}

func NewScheduler(scraper *scraper.Scraper, repo *database.Repository) *Scheduler {
    return &Scheduler{
        cron:    cron.New(),
        scraper: scraper,
        repo:    repo,
    }
}

func (s *Scheduler) Start(schedule string) error {
    _, err := s.cron.AddFunc(schedule, func() {
        log.Println("Starting scheduled scrape...")
        ctx := context.Background()
        if err := s.scraper.ScrapeAll(ctx); err != nil {
            log.Printf("Scheduled scrape failed: %v", err)
        }
    })

    if err != nil {
        return err
    }

    //runs every 2 hour
    _, err =s.cron.AddFunc("0 */2 * * *", func ()  {
        log.Println("Starting scheduled article cleanup...")
        ctx:= context.Background()
        if err := s.clearArticles(ctx); err != nil {
            log.Printf("Scheduled cleanup failed : %v", err)
        }
    })

    if err != nil {
        return err
    }

    s.cron.Start()
    log.Printf("Scheduler started with schedule: %s", schedule)
    log.Println("Article cleanup shceduled to run every two hours")
    return nil
}

func (s *Scheduler) clearArticles (ctx context.Context) error {
    log.Println("Clearing all articles from database...")
    if err := s.repo.ClearAllArticles(ctx); err != nil {
        return err
    }
    log.Println("Successfully cleared all articles")
    return nil
}

func (s *Scheduler) Stop() {
    s.cron.Stop()
}
