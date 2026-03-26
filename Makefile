# Absensholat API Makefile

.PHONY: all build run test test-coverage lint fmt clean docker-build docker-run migrate help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOMOD=$(GOCMD) mod
BINARY_NAME=absensholat-api

# Docker parameters
DOCKER_IMAGE=absensholat-api
DOCKER_TAG=latest

# Default target
all: test build

## build: Build the application
build:
	CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME) .

## run: Run the application
run:
	$(GOCMD) run Main.go

## test: Run all tests
test:
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-short: Run tests without verbose output
test-short:
	$(GOTEST) ./...

## lint: Run linters
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	$(GOFMT) ./...
	goimports -w .

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	$(GOCMD) clean

## deps: Download and tidy dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## swagger: Generate Swagger documentation
swagger:
	swag init -g Main.go -o docs

## docker-build: Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

## docker-run: Run Docker container
docker-run:
	docker-compose -f docker-compose.prod.yml up -d

## docker-stop: Stop Docker container
docker-stop:
	docker-compose -f docker-compose.prod.yml down

## docker-logs: View Docker container logs
docker-logs:
	docker-compose -f docker-compose.prod.yml logs -f

## migrate-up: Run database migrations (requires golang-migrate)
migrate-up:
	@echo "Running migrations..."
	migrate -path migrations -database "$(DATABASE_URL)" up

## migrate-down: Rollback database migrations
migrate-down:
	@echo "Rolling back migrations..."
	migrate -path migrations -database "$(DATABASE_URL)" down 1

## migrate-create: Create a new migration file
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

## security: Run security scan
security:
	gosec ./...
	govulncheck ./...

## install-tools: Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## dev: Run in development mode with hot reload (requires air)
dev:
	air

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
