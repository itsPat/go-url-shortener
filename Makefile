.PHONY: help run build db-up db-down db-shell clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Start the server
	go run ./cmd/server

build: ## Compile binary
	go build -o bin/server ./cmd/server

db-up: ## Start postgres
	docker compose up -d

db-down: ## Stop postgres
	docker compose down

db-shell: ## Open psql against the dev db
	docker compose exec db psql -U shortener -d shortener

clean: ## Remove build artifacts
	rm -rf bin/