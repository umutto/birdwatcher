.PHONY: proto lint test test-go test-python up down

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

lint:
	cd proto && buf lint

test: test-go test-python

test-go:
	go test $(GO_MODULES)

test-python:
	cd services/vision && uv run pytest

up:
	docker compose up -d

down:
	docker compose down
