CREATE TABLE IF NOT EXISTS sources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(512) NOT NULL,
    selector_title VARCHAR(255) NOT NULL,
    selector_link VARCHAR(255) NOT NULL,
    selector_summary VARCHAR(255),
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_url (url)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS articles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    source_id INT NOT NULL,
    title VARCHAR(512) NOT NULL,
    url VARCHAR(1024) NOT NULL,
    summary TEXT,
    scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (source_id) REFERENCES sources(id) ON DELETE CASCADE,
    UNIQUE KEY unique_article (url),
    INDEX idx_scraped_at (scraped_at),
    INDEX idx_source_id (source_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample sources
INSERT INTO sources (name, url, selector_title, selector_link, selector_summary) VALUES
('TechCrunch', 'https://techcrunch.com', '[class*="post"] a[href*="techcrunch"]', '[class*="post"] a[href*="techcrunch"]', 'div.post-block__content'),
('BBC News', 'https://www.bbc.com/news', 'a[data-testid="internal-link"]', 'a[data-testid="internal-link"]', 'p.gs-c-promo-summary'),
('The Guardian', 'https://www.theguardian.com/international', 'a[href*="/2025/"]', 'a[href*="/2025/"]', 'div[data-link-name*="article title"] p'),
('Reuters', 'https://www.reuters.com', 'a[href^="/world/"]' , 'a[href^="/world/"]' , 'p[data-testid="Text"]'),
('The Indian Express', 'https://indianexpress.com/', 'h3 a, h2 a', 'h3 a, h2 a', 'div.short-story p');
