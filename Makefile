.PHONY: build run test lint clean dev web migrate-up migrate-down

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

dev:
	docker-compose -f deployments/docker/docker-compose.yml up --build

web:
	cd apps/web && npm run dev

migrate-up:
	go run ./cmd/initdb

migrate-down:
	go run ./cmd/initdb down
