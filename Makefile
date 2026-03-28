APP_NAME := eth-call
BUILD_DIR := bin
GO := go

.PHONY: all build run test lint fmt clean help

all: build

build: ## Build the application
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

run: build ## Run the application
	./$(BUILD_DIR)/$(APP_NAME)

test: ## Run tests
	$(GO) test -v -race ./...

test-coverage: ## Run tests with coverage
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	$(GO) fmt ./...
	goimports -w .

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

mod-tidy: ## Tidy go modules
	$(GO) mod tidy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
