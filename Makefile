# Every target runs in a container; nothing but Docker is needed on the host.
COMPOSE := docker compose
GO      := $(COMPOSE) run --rm --no-deps go
GO_DB   := $(COMPOSE) run --rm go

# gotestsum prints one line per test and a DONE summary; plain `go test` only
# says "ok <package>". Override with FORMAT=pkgname for one line per package.
GOTESTSUM := gotest.tools/gotestsum@v1.13.0
FORMAT    ?= testname

.DEFAULT_GOAL := help
.PHONY: help up down restart logs build test test-integration cover lint vet fmt tidy check psql clean

help: ## Show this help
	@grep -hE '^[a-z-]+:.*##' $(MAKEFILE_LIST) | sed 's/:.*##/\t/' | expand -t22

up: ## Build and start the stack
	$(COMPOSE) up --build -d

down: ## Stop the stack (make down ARGS=-v also drops the database)
	$(COMPOSE) down $(ARGS)

restart: down up ## Restart the stack

logs: ## Follow the api logs
	$(COMPOSE) logs -f api

build: ## Compile the binary
	$(GO) build ./...

test: ## Unit and API tests, no database
	$(GO) run $(GOTESTSUM) --format $(FORMAT) -- ./test/...

test-integration: ## All tests, including the PostgreSQL suite
	$(GO_DB) run $(GOTESTSUM) --format $(FORMAT) -- -tags=integration -count=1 ./test/...

cover: ## Write coverage.out and print the summary
	$(GO_DB) test -tags=integration -coverpkg=./internal/... -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1

lint: ## Run staticcheck over every build tag
	$(GO) run honnef.co/go/tools/cmd/staticcheck@latest -tags=integration ./...

vet: ## Run go vet over every build tag
	$(GO) vet -tags=integration ./...

fmt: ## Format the sources
	$(GO) fmt ./...

tidy: ## Tidy go.mod and go.sum
	$(GO) mod tidy

check: fmt vet lint test ## Format, vet, lint and test

psql: ## Open a psql shell on the database
	$(COMPOSE) exec postgres psql -U $${POSTGRES_USER:-myuser} -d $${POSTGRES_DB:-alarm}

clean: ## Stop the stack and remove volumes, images and coverage output
	$(COMPOSE) down -v --rmi local
	rm -f coverage.out
