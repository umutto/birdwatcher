.PHONY: env proto sqlc migrate migrate-reset seed lint lint-proto lint-go lint-python lint-sql lint-web format format-go format-sql format-python format-web test test-go test-python verify setup-hooks media up down dev-catalog dev-frontier dev-scraper dev-textparser dev-vision dev-web tools

MEDIA_TAG := media-v2
MEDIA_ARCHIVE := birdwatcher-$(MEDIA_TAG).tar.gz
MEDIA_URL := https://github.com/umutto/birdwatcher/releases/download/$(MEDIA_TAG)/$(MEDIA_ARCHIVE)
MEDIA_SHA256 := c5ab5e21ffb6ae1dbca42bbc46692f4e15702a7c8b114e6a235015b67f7129b2

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

env:
	@for src in .env.example services/*/.env.example; do \
		dst="$${src%.example}"; \
		if [ -f "$$dst" ]; then \
			echo "exists:  $$dst"; \
		else \
			cp "$$src" "$$dst"; \
			echo "created: $$dst"; \
		fi; \
	done

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

lint: lint-proto lint-go lint-python lint-sql lint-web

lint-proto:
	cd proto && buf lint

lint-go:
	golangci-lint run $(GO_MODULES)

lint-python:
	cd services/vision && uv run ruff check .

lint-sql:
	cd services/catalog && sqlc compile
	sqruff lint services/catalog

lint-web:
	cd web && pnpm lint

format: format-go format-python format-sql format-web

format-go:
	golangci-lint fmt services shared

format-sql:
	sqruff fix services/catalog

format-python:
	cd services/vision && uv run ruff format .

format-web:
	cd web && pnpm format

test: test-go test-python

test-go:
	go test -race $(GO_MODULES)

test-python:
	cd services/vision && uv run pytest

verify:
	@./.githooks/pre-push

setup-hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-push

tools:
	go install github.com/bufbuild/buf/cmd/buf@v1.75.0
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
	go install github.com/air-verse/air@latest
	uv tool install sqruff==0.44.0

media:
	@set -eu; \
	if [ -d media/images ] && [ -d media/audio ] && [ -f media/.version ] && [ "$$(cat media/.version)" = "$(MEDIA_TAG)" ] && [ "$(FORCE)" != "1" ]; then \
		echo "Media assets ($(MEDIA_TAG)) already present in ./media (run 'make media FORCE=1' to re-download)"; \
		exit 0; \
	fi; \
	tmp=$$(mktemp); \
	trap 'rm -f "$$tmp"' EXIT INT TERM; \
	echo "Downloading $(MEDIA_ARCHIVE) ($(MEDIA_TAG))..."; \
	curl -fL --progress-bar -o "$$tmp" "$(MEDIA_URL)"; \
	echo "Verifying SHA-256 checksum..."; \
	if command -v sha256sum >/dev/null 2>&1; then \
		printf '%s  %s\n' "$(MEDIA_SHA256)" "$$tmp" | sha256sum -c -; \
	else \
		printf '%s  %s\n' "$(MEDIA_SHA256)" "$$tmp" | shasum -a 256 -c -; \
	fi; \
	rm -rf media/images media/audio; \
	mkdir -p media; \
	echo "Extracting into ./media/..."; \
	tar -xzf "$$tmp" -C media; \
	printf '%s\n' "$(MEDIA_TAG)" > media/.version; \
	echo "Media assets ($(MEDIA_TAG)) ready in ./media/"

up:
	docker compose up -d --wait

down:
	docker compose down

dev-catalog:
	cd services/catalog && air

dev-frontier:
	cd services/frontier && set -a && . ./.env && set +a && go run .

dev-scraper:
	cd services/scraper && set -a && . ./.env && set +a && go run .

dev-textparser:
	cd services/textparser && set -a && . ./.env && set +a && go run .

dev-vision:
	cd services/vision && uv run --env-file .env vision

dev-web:
	cd web && pnpm dev
