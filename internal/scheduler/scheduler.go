package scheduler

import (
    "context"
    "log"

    "github.com/robfig/cron/v3"
    "news-scraper/internal/scraper"
)

type Scheduler struct {
    cron    *cron.Cron
    scraper *scraper.Scraper
}

func NewScheduler(scraper *scraper.Scraper) *Scheduler {
    return &Scheduler{
        cron:    cron.New(),
        scraper: scraper,
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

    s.cron.Start()
    log.Printf("Scheduler started with schedule: %s", schedule)
    return nil
}

func (s *Scheduler) Stop() {
    s.cron.Stop()
}
