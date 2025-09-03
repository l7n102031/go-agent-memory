.PHONY: help test build clean install deps lint

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Install dependencies
	go mod download
	go mod tidy

build: ## Build the package
	go build -v ./...

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./... ./tests/...

coverage: test ## Show test coverage
	go tool cover -html=coverage.out

lint: ## Run linters
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

example: ## Run the example
	cd examples && go run integration.go

docker-deps: ## Start Redis and PostgreSQL for testing
	cd deployment && docker-compose up -d

docker-stop: ## Stop test dependencies
	cd deployment && docker-compose down

install: ## Install the package locally
	go install ./...

clean: ## Clean build artifacts
	rm -f coverage.out
	go clean -cache
