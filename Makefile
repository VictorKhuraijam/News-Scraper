.PHONY: build run test migrate clean

build:
	go build -o bin/server cmd/server/main.go

tg:
	templ generate

run:
	go run cmd/server/main.go

test:
	go test -v ./...

migrate:
	mysql -u root -p news_scraper < migrations/001_init.sql

clean:
	rm -rf bin/

deps:
	go mod download
	go mod tidy

.DEFAULT_GOAL := run
