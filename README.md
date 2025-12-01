# News Scraper - GOTTH Stack

A concurrent web scraper for news sites built with Go, HTMX, Tailwind CSS, and Templ.

## Features

- **Concurrent Scraping**: Uses goroutines and channels for parallel processing
- **Rate Limiting**: Token bucket algorithm to respect website limits
- **Context Management**: Proper timeout and cancellation handling
- **Scheduled Scraping**: Automatic periodic execution with cron
- **Dynamic UI**: HTMX for seamless updates without page reloads
- **MySQL Storage**: Persistent storage with proper indexing
- **Multiple Sources**: Support for multiple news websites

## Tech Stack

- **Backend**: Go 1.21+ with Fiber framework
- **Database**: MySQL 8.0+
- **Frontend**: Templ templates, HTMX, Tailwind CSS
- **Scheduling**: Cron-based automatic scraping
- **Parsing**: goquery for HTML parsing

## Project Structure

```
news-scraper/
├── cmd/server/          # Application entry point
├── configs/             # Configuration files
├── internal/
│   ├── scraper/        # Scraping logic
│   ├── models/         # Data models
│   ├── database/       # Database layer
│   ├── handlers/       # HTTP handlers
│   └── scheduler/      # Cron scheduler
├── web/
│   ├── templates/      # Templ templates
│   └── static/         # CSS/JS assets
└── migrations/         # SQL migrations
```

## Setup Instructions

### Prerequisites

- Go 1.21 or higher
- MySQL 8.0 or higher
- Templ CLI (`go install github.com/a-h/templ/cmd/templ@latest`)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd news-scraper
```

2. Install dependencies:
```bash
make deps
```

3. Create database:
```bash
mysql -u root -p -e "CREATE DATABASE news_scraper CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

4. Run migrations:
```bash
make migrate
```

5. Configure application:
```bash
cp configs/.env.example configs/.env
# Edit configs/config.yaml with your settings
```

6. Generate Templ templates:
```bash
templ generate
```

7. Run the application:
```bash
make run
```

The server will start on `http://localhost:3000`

## Configuration

Edit `configs/config.yaml`:

```yaml
server:
  port: 3000
  host: 0.0.0.0

database:
  host: localhost
  port: 3306
  user: root
  password: your_password
  name: news_scraper

scraper:
  workers: 5              # Number of concurrent workers
  timeout: 30s            # HTTP request timeout
  rate_limit: 10          # Requests per second
  user_agent: "NewsBot/1.0"
  schedule: "0 */6 * * *" # Cron schedule (every 6 hours)
```

## Adding News Sources

Add sources to the database:

```sql
INSERT INTO sources (name, url, selector_title, selector_link, selector_summary)
VALUES (
    'Source Name',
    'https://example.com',
    'h2.article-title',      -- CSS selector for title
    'a.article-link',         -- CSS selector for link
    'p.article-summary'       -- CSS selector for summary (optional)
);
```

## API Endpoints

- `GET /` - Home page
- `GET /articles` - Articles page
- `GET /api/articles` - Get recent articles (JSON)
- `GET /api/articles/source/:sourceId` - Get articles by source (JSON)
- `POST /api/scrape` - Trigger manual scrape

## Development

### Build
```bash
make build
```

### Run tests
```bash
make test
```

### Clean build artifacts
```bash
make clean
```

## Security Considerations

1. **Rate Limiting**: Implemented to avoid overwhelming target websites
2. **User Agent**: Configurable to identify your bot
3. **Robots.txt**: Respect website's robots.txt (not auto-checked, manual review needed)
4. **Terms of Service**: Ensure compliance with each website's ToS
5. **Error Handling**: Graceful failure handling to avoid crashes

## Concurrency Features

- **Worker Pool**: Configurable number of goroutines
- **Channel Communication**: Job distribution via channels
- **Context Cancellation**: Proper cleanup on shutdown
- **WaitGroups**: Coordinated goroutine completion
- **Rate Limiter**: Thread-safe token bucket implementation

## Future Enhancements

- [ ] Add proxy support
- [ ] Implement user authentication
- [ ] Add article filtering by topic
- [ ] Support for RSS feeds
- [ ] Export articles to JSON/CSV
- [ ] Add search functionality
- [ ] Implement article deduplication
- [ ] Add notification system

## License

MIT License

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
```

This complete implementation provides:

1. **Proper concurrency**: Worker pools, goroutines, channels, and context
2. **Rate limiting**: Token bucket algorithm
3. **Full GOTTH stack**: Go, HTMX, Tailwind, Templ
4. **MySQL integration**: With proper migrations
5. **Scheduled scraping**: Cron-based automation
6. **Clean architecture**: Separated concerns with proper folder structure
7. **Security features**: Rate limiting, user agent configuration
8. **Production-ready**: Error handling, logging, graceful shutdown

To use this:
1. Generate templ templates: `templ generate`
2. Set up MySQL database
3. Run migrations
4. Configure settings
5. Start the server: `make run`
