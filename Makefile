.PHONY: help build run stop clean fmt docker-build docker-up docker-down docker-logs

# project variables
PROJECT_NAME = pr-reviewer-service
BINARY_NAME = pr-reviewer-service

help:
	@echo Available commands:
	@echo.
	@echo Development:
	@echo   build          - build the application binary
	@echo   run            - run the application locally
	@echo   clean          - remove build artifacts
	@echo   fmt            - format code with gofmt
	@echo.
	@echo Docker:
	@echo   docker-build   - build docker image
	@echo   docker-up      - start services with docker-compose
	@echo   docker-down    - stop and remove containers
	@echo   docker-down-v  - stop containers and remove volumes
	@echo   docker-logs    - show logs from all containers
	@echo   docker-logs-app - show logs from app container only
	@echo   docker-restart - restart services
	@echo   docker-ps      - show running containers
	@echo.
	@echo Database:
	@echo   db-shell       - connect to postgres shell
	@echo.
	@echo Utilities:
	@echo   deps           - download dependencies
	@echo   verify         - verify dependencies
	@echo.
	@echo Quick Start:
	@echo   dev            - quick start for development (clean slate)
	@echo   all            - run all checks and build

build:
	@echo Building $(BINARY_NAME)...
	go build -ldflags="-w -s" -o bin/$(BINARY_NAME).exe ./cmd/server

run:
	@echo Running $(PROJECT_NAME)...
	go run ./cmd/server

clean:
	@echo Cleaning build artifacts...
	@if exist bin rmdir /s /q bin
	@if exist vendor rmdir /s /q vendor
	@if exist coverage.out del coverage.out
	@if exist coverage.html del coverage.html

fmt:
	@echo Formatting code...
	gofmt -s -w .

docker-build:
	@echo Building docker image...
	docker build -t $(PROJECT_NAME):latest .

docker-up:
	@echo Starting services...
	docker-compose up -d
	@echo.
	@echo Services started. Application available at http://localhost:8080
	@echo Health check: http://localhost:8080/health
	@echo.
	@echo Run 'make docker-logs' to see logs

docker-down:
	@echo Stopping services...
	docker-compose down

docker-down-v:
	@echo Stopping services and removing volumes...
	docker-compose down -v

docker-logs:
	docker-compose logs -f

docker-logs-app:
	docker-compose logs -f app

docker-restart:
	@echo Restarting services...
	docker-compose restart

docker-ps:
	docker-compose ps

db-shell:
	@echo Connecting to database...
	docker-compose exec postgres psql -U postgres -d postgres

deps:
	@echo Downloading dependencies...
	go mod download
	go mod tidy

verify:
	@echo Verifying dependencies...
	go mod verify

dev: docker-down-v docker-up
	@echo.
	@echo Development environment ready!
	@echo Application: http://localhost:8080
	@echo Health check: http://localhost:8080/health

all: clean deps fmt docker-build
	@echo.
	@echo All checks passed and build complete!