.PHONY: help test test-race lint build run cover mocks migrate-up migrate-down ci web-build web-dev web-test web-test-watch clean

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

MOCKGEN ?= go run go.uber.org/mock/mockgen

mocks: ## Regenerate gomock mocks
	$(MOCKGEN) -source=internal/domain/ports.go -destination=internal/persistence/mocks/mock_repos.go -package=mocks

migrate-up: ## Apply pending migrations to $$DB_PATH
	$(GO) run ./cmd/agent --migrate-up

migrate-down: ## Roll the latest migration back on $$DB_PATH
	$(GO) run ./cmd/agent --migrate-down

web-build: ## Build the embedded frontend and stage it for go:embed
	cd web && npm install && npm run build
	@rm -rf internal/web/dist
	@mkdir -p internal/web/dist
	@cp -R web/dist/. internal/web/dist/
	@touch internal/web/dist/.gitkeep

web-dev: ## Run frontend dev server (Vite, HMR)
	cd web && npm run dev

web-test: ## Run frontend tests one-shot
	cd web && npm test -- --run

web-test-watch: ## Run frontend tests in watch mode
	cd web && npm test

ci: lint test build ## Run the full local CI check

clean: ## Remove build artifacts
	rm -rf bin/ $(COVERAGE)
