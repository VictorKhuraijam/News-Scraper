package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "gopkg.in/yaml.v3"

    "news-scraper/internal/database"
    "news-scraper/internal/handlers"
    "news-scraper/internal/scheduler"
    "news-scraper/internal/scraper"
)

type Config struct {
    Server struct {
        Port string `yaml:"port"`
        Host string `yaml:"host"`
    } `yaml:"server"`
    Database struct {
        Host     string `yaml:"host"`
        Port     int    `yaml:"port"`
        User     string `yaml:"user"`
        Password string `yaml:"password"`
        Name     string `yaml:"name"`
    } `yaml:"database"`
    Scraper struct {
        Workers   int    `yaml:"workers"`
        Timeout   string `yaml:"timeout"`
        RateLimit int    `yaml:"rate_limit"`
        UserAgent string `yaml:"user_agent"`
        Schedule  string `yaml:"schedule"`
    } `yaml:"scraper"`
}

func loadConfig() (*Config, error) {
    data, err := os.ReadFile("configs/config.yaml")
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

func main() {
    // Load configuration
    cfg, err := loadConfig()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // Connect to database
    db, err := database.Connect(database.Config{
        Host:     cfg.Database.Host,
        Port:     cfg.Database.Port,
        User:     cfg.Database.User,
        Password: cfg.Database.Password,
        Name:     cfg.Database.Name,
    })
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    log.Println("Database connected successfully")

    // Initialize repository
    repo := database.NewRepository(db)

    // Parse timeout
    timeout, err := time.ParseDuration(cfg.Scraper.Timeout)
    if err != nil {
        timeout = 30 * time.Second
    }

    // Initialize scraper
    scraperInstance := scraper.NewScraper(repo, scraper.Config{
        Workers:   cfg.Scraper.Workers,
        Timeout:   timeout,
        RateLimit: cfg.Scraper.RateLimit,
        UserAgent: cfg.Scraper.UserAgent,
    })

    // Initialize scheduler
    sched := scheduler.NewScheduler(scraperInstance)
    if err := sched.Start(cfg.Scraper.Schedule); err != nil {
        log.Printf("Warning: Failed to start scheduler: %v", err)
    }
    defer sched.Stop()

    // Initialize handlers
    homeHandler := handlers.NewHomeHandler(repo)
    articlesHandler := handlers.NewArticlesHandler(repo)
    scrapeHandler := handlers.NewScrapeHandler(scraperInstance)

    // Create Fiber app
    app := fiber.New(fiber.Config{
        Views: nil, // Use templ for rendering
    })

    // Middleware
    app.Use(logger.New())
    app.Use(recover.New())

    // Static files
    app.Static("/static", "./web/static")

    // Routes
    app.Get("/", homeHandler.Index)
    app.Get("/articles", articlesHandler.RenderArticles)

    // API routes
    api := app.Group("/api")
    api.Get("/articles", articlesHandler.GetRecent)
    api.Get("/articles/source/:sourceId", articlesHandler.GetBySource)
    api.Post("/scrape", scrapeHandler.TriggerScrape)

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-quit
        log.Println("Shutting down server...")

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        if err := app.ShutdownWithContext(ctx); err != nil {
            log.Fatal("Server forced to shutdown:", err)
        }
    }()

    // Start server
    addr := cfg.Server.Host + ":" + cfg.Server.Port
    log.Printf("Server starting on %s", addr)

    if err := app.Listen(addr); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}
