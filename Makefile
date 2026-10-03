.PHONY: proto sqlc lint lint-proto lint-go lint-python lint-sql format-python test test-go test-python verify setup-hooks up down

GO_MODULES := \
	./services/catalog/... \
	./services/frontier/... \
	./services/scraper/... \
	./services/textparser/... \
	./shared/logger/... \
	./shared/nats/... \
	./shared/telemetry/... \
	./gen/go/...

proto:
	cd proto && buf generate

sqlc:
	cd services/catalog && sqlc generate

lint: lint-proto lint-go lint-python lint-sql

lint-proto:
	cd proto && buf lint

lint-go:
	golangci-lint run $(GO_MODULES)

lint-python:
	cd services/vision && uv run ruff check .

lint-sql:
	cd services/catalog && sqlc compile

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
