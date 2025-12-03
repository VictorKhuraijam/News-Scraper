package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
    Host     string
    Port     int
    User     string
    Password string
    Name     string
}

func Connect(cfg Config) (*sql.DB, error) {
    //allows for multi statement query //multiStatements=true
    //connection to sql server
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?multiStatements=true&parseTime=true&charset=utf8mb4",
        cfg.User, cfg.Password, cfg.Host, cfg.Port)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    defer db.Close()

    //Create database

    _, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", cfg.Name))
    if err != nil{
        log.Fatalf("Error in creating database: %v",err)
    }
    fmt.Printf("Database '%s' ensured to exist.\n", cfg.Name)
    db.Close()

    //re-connect to the newly created database

    dbDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true&parseTime=true&charset=utf8mb4",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
    db, err = sql.Open("mysql", dbDsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    defer db.Close()

    createTable(db)

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)

    return db, nil
}


func createTable(db *sql.DB){
//multi statement query
    query := `
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
        url VARCHAR(768) NOT NULL,
        summary TEXT,
        scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (source_id) REFERENCES sources(id) ON DELETE CASCADE,
        UNIQUE KEY unique_article (url),
        INDEX idx_scraped_at (scraped_at),
        INDEX idx_source_id (source_id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
    `

    _,err := db.Exec(query)
    if err != nil{
        log.Fatal("Failed to create table:", err)
    }
}
