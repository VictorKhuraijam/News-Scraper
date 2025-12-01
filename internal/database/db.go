package database

import (
    "database/sql"
    "fmt"
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
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)

    return db, nil
}
