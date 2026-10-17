.PHONY: proto sqlc migrate migrate-reset seed lint lint-proto lint-go lint-python lint-sql format format-go format-sql format-python test test-go test-python verify setup-hooks up down watch-catalog dev

GO_MODULES := \
	./services/catalog/... \
	./services/frontier/... \
	./services/scraper/... \
	./services/textparser/... \
	./shared/env/... \
	./shared/logger/... \
	./shared/nats/... \
	./shared/telemetry/... \
	./gen/go/...

proto:
	cd proto && buf generate

sqlc:
	cd services/catalog && sqlc generate

migrate:
	cd services/catalog && set -a && . ./.env && set +a && goose -dir migrations postgres "$$DATABASE_URL" up

migrate-reset:
	cd services/catalog && set -a && . ./.env && set +a && goose -dir migrations postgres "$$DATABASE_URL" reset

seed:
	cd services/catalog && set -a && . ./.env && set +a && go run ./cmd/seed

lint: lint-proto lint-go lint-python lint-sql

lint-proto:
	cd proto && buf lint

lint-go:
	golangci-lint run $(GO_MODULES)

lint-python:
	cd services/vision && uv run ruff check .

lint-sql:
	cd services/catalog && sqlc compile
	sqruff lint services/catalog

format: format-go format-python format-sql

format-go:
	gofmt -w services shared

format-sql:
	sqruff fix services/catalog

format-python:
	cd services/vision && uv run ruff format .

test: test-go test-python

test-go:
	go test $(GO_MODULES)

test-python:
	cd services/vision && uv run pytest

verify:
	@./.githooks/pre-push

setup-hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-push

up:
	docker compose up -d

down:
	docker compose down

watch-catalog:
	cd services/catalog && air

dev: watch-catalog
