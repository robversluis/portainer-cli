.PHONY: help build test lint clean install run dev fmt vet tidy

# Variables
BINARY_NAME=portainer-cli
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/portainer-cli

build-all: ## Build for all platforms
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/portainer-cli
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/portainer-cli
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/portainer-cli
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/portainer-cli
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/portainer-cli

test: ## Run tests
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-coverage: test ## Run tests with coverage report
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

fmt: ## Format code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

tidy: ## Tidy go modules
	$(GOMOD) tidy

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.txt coverage.html

install: build ## Install the binary to GOPATH/bin
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

run: build ## Build and run the binary
	./bin/$(BINARY_NAME)

dev: ## Run in development mode
	$(GOCMD) run ./cmd/portainer-cli

deps: ## Download dependencies
	$(GOMOD) download

verify: fmt vet lint test ## Run all verification steps

release: ## Create a release (requires goreleaser)
	@which goreleaser > /dev/null || (echo "goreleaser not installed. Install from https://goreleaser.com/install/" && exit 1)
	goreleaser release --clean

snapshot: ## Create a snapshot release
	@which goreleaser > /dev/null || (echo "goreleaser not installed. Install from https://goreleaser.com/install/" && exit 1)
	goreleaser release --snapshot --clean
