.PHONY: proto lint lint-proto lint-go lint-python format-python test test-go test-python up down

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

lint: lint-proto lint-go lint-python

lint-proto:
	cd proto && buf lint

lint-go:
	golangci-lint run $(GO_MODULES)

lint-python:
	cd services/vision && uv run ruff check .

format-python:
	cd services/vision && uv run ruff format .

test: test-go test-python

test-go:
	go test $(GO_MODULES)

test-python:
	cd services/vision && uv run pytest

up:
	docker compose up -d

down:
	docker compose down
