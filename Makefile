.PHONY: help test test-race lint build run cover mocks migrate ci web-build clean

GO ?= go
BIN := bin/agent
CMD := ./cmd/agent
COVERAGE := coverage.out

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "%-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run all tests
	$(GO) test ./...

test-race: ## Run all tests with the race detector
	$(GO) test -race ./...

cover: ## Run tests with coverage report
	$(GO) test -coverprofile=$(COVERAGE) ./...
	$(GO) tool cover -func=$(COVERAGE)

lint: ## Run golangci-lint
	golangci-lint run ./...

build: ## Build the agent binary
	$(GO) build -o $(BIN) $(CMD)

run: build ## Build and run the agent
	./$(BIN)

mocks: ## Regenerate gomock mocks (Phase 1+)
	$(GO) generate ./...

migrate: ## Run database migrations (wired in Phase 1, no-op for now)
	@echo "migrate target reserved for Phase 1 (goose). Skipping."

web-build: ## Build the embedded frontend (Phase 4+)
	cd web && npm install && npm run build

ci: lint test build ## Run the full local CI check

clean: ## Remove build artifacts
	rm -rf bin/ $(COVERAGE)
