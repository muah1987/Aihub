.PHONY: build run test lint migrate-up migrate-down docker-up docker-down web-dev web-build clean

# Go binary
BINARY_NAME=aihub-server
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -v

lint:
	golangci-lint run ./...

# Database migrations
migrate-up:
	go run ./cmd/server -migrate up

migrate-down:
	go run ./cmd/server -migrate down

# Docker
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose build

# Frontend
web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-install:
	cd web && npm install

# Full setup
setup: web-install
	go mod tidy
	@echo "Setup complete. Copy .env.example to .env and configure."

clean:
	rm -rf $(BUILD_DIR) web/dist web/node_modules
