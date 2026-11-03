# birdwatcher

[![CI](https://github.com/umutto/birdwatcher/actions/workflows/ci.yml/badge.svg)](https://github.com/umutto/birdwatcher/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Python / uv](https://img.shields.io/badge/uv-Python_3.14+-DE5FE9?logo=astral&logoColor=white)](https://docs.astral.sh/uv/)
[![Vue 3](https://img.shields.io/badge/Vue-3-4FC08D?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![Protobuf](https://img.shields.io/badge/Protobuf-Buf_v2-244c5a?logo=buf&logoColor=white)](https://buf.build/)
[![NATS JetStream](https://img.shields.io/badge/NATS-JetStream-27AAE1?logo=nats.io&logoColor=white)](https://nats.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**birdwatcher** is a distributed, event-driven web crawler and bird recognition pipeline built with Go, Python, and Vue 3. Given a starting URL, it crawls web pages, extracts text and images, detects bird species through text matching and computer vision, and indexes every sighting against the 435 plates of John James Audubon's *The Birds of America*.

---

## How It Works

```mermaid
flowchart TD
    SPA["web (Vue 3 Dashboard)"] <-->|HTTP REST| CAT["catalog (Go API & CLI)"]
    CAT <-->|SQL| DB[(PostgreSQL 18)]
    CAT -->|gRPC: StartJob| FR["frontier (Go URL Queue & Rate Limiter)"]

    FR -->|PageFetchRequested| NATS{{NATS JetStream}}
    NATS -.->|PageFetchRequested| SC["scraper (Go HTML Scraper)"]
    SC -->|UrlDiscovered| NATS
    NATS -.->|UrlDiscovered| FR
    SC -->|PageScraped| NATS

    NATS -.->|PageScraped| TP["textparser (Go Text Matcher)"]
    NATS -.->|PageScraped| VS["vision (Python Image Classifier)"]

    TP -->|SightingText| NATS
    VS -->|SightingImage| NATS
    FR -->|JobCompleted| NATS

    NATS -.->|Sightings & Job Status| CAT
```

> **Diagram key**: Solid lines (`-->`) show direct requests and published events; dashed lines (`-.->`) show workers subscribing to events from NATS JetStream.

1. **Submit a Job**: A user submits a URL through the Vue dashboard (`web`) or CLI to `catalog` (`POST /api/v1/crawls`).
2. **Schedule URLs**: `catalog` records the crawl in PostgreSQL and calls `frontier` over gRPC (`FrontierService.StartJob`). `frontier` tracks visited links, enforces domain rate limits, and publishes `PageFetchRequested` events to NATS JetStream.
3. **Scrape Content**: `scraper` workers fetch the HTML page, send discovered `<a href>` links back to `frontier` via `UrlDiscovered`, and publish the page title, text, and `<img src>` URLs via `PageScraped`.
4. **Detect Birds**: Two services process `PageScraped` in parallel:
   - `textparser` uses Aho-Corasick string matching and text normalization to scan page text for bird common names, scientific names, and aliases, emitting `SightingText`.
   - `vision` downloads page images, classifies birds using a computer vision model, and emits `SightingImage`.
5. **Index Results**: `catalog` consumes sighting and completion events, saves observations to PostgreSQL, and serves crawl progress and the species catalog to the frontend and CLI.

---

## Services & Packages

| Directory | Stack | Port | Purpose |
| :--- | :--- | :--- | :--- |
| `services/catalog` | Go (`net/http`, `pgx`, `sqlc`) | `8080` | Owns the PostgreSQL database, records sightings from NATS, and exposes two interfaces: an HTTP API backend for `web` (`cmd/catalog`) and a CLI (`cmd/cli`). |
| `services/frontier` | Go (`gRPC`, NATS) | `9090` | Breadth-first URL queue, visited-link deduplication, `robots.txt` checks, and host rate limiting. |
| `services/scraper` | Go (`net/http`, `goquery`) | — | Stateless worker that fetches HTML and extracts links, text, and image URLs. |
| `services/textparser` | Go (Aho-Corasick, NLP) | — | Worker that scans scraped text for bird names and aliases using Aho-Corasick matching and NLP. |
| `services/vision` | Python 3.14+ (`uv`) | — | Worker that downloads scraped images and runs bird species classification. |
| `web` | Vue 3, TypeScript, Vite, Tailwind v4 | `5173` | Single-page app for browsing the 435-species catalog and monitoring crawl jobs. |
| `proto` & `gen` | Protobuf, Buf v2 | — | Shared gRPC and event schemas (`events.proto`, `frontier.proto`) and generated Go/Python code. |
| `shared/*` | Go | — | Internal packages for config loading (`env`), structured logging (`logger`), NATS (`nats`), and tracing (`telemetry`). |

### NATS JetStream Events

All async events use Protobuf binary encoding on the `BIRDWATCHER` stream (`birdwatcher.v1.>`):

| Subject | Message | Producer | Consumer |
| :--- | :--- | :--- | :--- |
| `birdwatcher.v1.page.fetch` | `PageFetchRequested` | `frontier` | `scraper` |
| `birdwatcher.v1.url.discovered` | `UrlDiscovered` | `scraper` | `frontier` |
| `birdwatcher.v1.page.scraped` | `PageScraped` | `scraper` | `textparser`, `vision`, `catalog` |
| `birdwatcher.v1.sighting.text` | `SightingText` | `textparser` | `catalog` |
| `birdwatcher.v1.sighting.image` | `SightingImage` | `vision` | `catalog` |
| `birdwatcher.v1.job.completed` | `JobCompleted` | `frontier` | `catalog` |

---

## Prerequisites

Make sure the following tools are installed on your system (Linux / macOS / WSL2):

- **Go** `1.27+`
- **Python** `3.14+` with [**uv**](https://docs.astral.sh/uv/)
- **Node.js** `24+` with [**pnpm**](https://pnpm.io/)
- **Docker** & **Docker Compose**

Install the project CLI tools:

```bash
# Protobuf compiler & linter
go install github.com/bufbuild/buf/cmd/buf@latest

# Type-safe SQL generator & database migration runner
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/pressly/goose/v3/cmd/goose@latest

# Go linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Optional: live-reload runner for local Go development
go install github.com/air-verse/air@latest

# SQL linter & formatter
uv tool install sqruff
```

---

## Getting Started

### 1. Configure Environment Files

Copy the example environment files for Docker Compose and each service:

```bash
cp .env.example .env
for svc in catalog frontier scraper textparser vision; do
  cp "services/$svc/.env.example" "services/$svc/.env"
done
```

### 2. Start Local Infrastructure

Start **NATS JetStream** (`localhost:4222`, monitoring at `http://localhost:8222`) and **PostgreSQL 18** (`localhost:5432`):

```bash
make up
```

### 3. Run Migrations & Seed the Catalog

Apply the database schema and populate the 435 Audubon bird species:

```bash
make migrate
make seed
```

### 4. Set Up Git Hooks

Enable the pre-push hook to check for `sqlc`, `buf`, and `go.work.sum` drift before pushing:

```bash
make setup-hooks
```

### 5. Run the Application

```bash
# Run Catalog API with live reload (http://localhost:8080)
make dev

# Run Go workers (in separate terminals)
go run ./services/frontier
go run ./services/scraper
go run ./services/textparser

# Run Python vision worker
cd services/vision && uv run vision

# Run Vue 3 frontend (http://localhost:5173)
cd web && pnpm install && pnpm dev
```

---

## Development Commands

Run tasks from the repository root using `make`:

| Command | Description |
| :--- | :--- |
| `make up` / `make down` | Start or stop NATS JetStream and PostgreSQL containers. |
| `make migrate` / `make migrate-reset` | Apply or roll back PostgreSQL migrations with `goose`. |
| `make seed` | Seed the 435 Audubon species into PostgreSQL. |
| `make proto` | Generate Go and Python stubs from `proto/` using `buf`. |
| `make sqlc` | Generate Go database queries in `services/catalog` using `sqlc`. |
| `make lint` | Run all linters (`buf`, `golangci-lint`, `ruff`, `sqlc`, `sqruff`). |
| `make format` | Format Go (`gofmt`), Python (`ruff`), and SQL (`sqruff`) files. |
| `make test` | Run Go and Python test suites. |
| `make verify` | Run pre-push code generation and workspace drift checks manually. |
| `make dev` | Run `services/catalog` with `air` hot-reloading. |

> **Go Workspace Note**: This repository uses `go.work` to link local Go modules (`gen/go`, `shared/*`, `services/*`). Run `go work sync` at the root to sync dependencies instead of running `go mod tidy` inside individual modules.

---

## Catalog HTTP API

`services/catalog` serves the following endpoints on port `8080`:

| Method & Path | Description |
| :--- | :--- |
| `GET /livez`, `GET /healthz` | Liveness check (`200 OK`). |
| `GET /readyz` | Readiness check (verifies PostgreSQL connectivity). |
| `GET /api/v1/species` | List all 435 bird species. |
| `GET /api/v1/species/dictionary` | Lightweight list of species names, scientific names, and aliases for text matching. |
| `GET /api/v1/species/{plate}` | Get details for a single species by Audubon plate number (`1–435`). |
| `GET /api/v1/species/{plate}/sightings` | List paginated sightings for a specific species. |
| `POST /api/v1/crawls` | Create a new crawl job and dispatch it to `frontier`. |
| `GET /api/v1/crawls` | List paginated crawl jobs with sighting summaries. |
| `GET /api/v1/crawls/{id}` | Get details and progress counters for a crawl job. |
| `GET /api/v1/crawls/{id}/catalog` | List all 435 species with sighting counts for a specific crawl. |
| `GET /api/v1/crawls/{id}/sightings` | List paginated sightings discovered during a crawl. |

---

## CI/CD & Deployment

- **Continuous Integration (`.github/workflows/ci.yml`)**: Path-filtered GitHub Actions jobs verify Protobuf drift (`buf`), SQL generation and linting (`sqlc`, `sqruff`), Go linting and race-enabled tests (`golangci-lint`, `go test -race`), Python checks (`ruff`, `pytest`), and frontend builds (`oxlint`, `eslint`, `oxfmt`, `vue-tsc`, `vite`).
- **Container Registry**: Pushes to `main` build multi-stage Docker images for modified services and publish them to GitHub Container Registry (`ghcr.io/umutto/birdwatcher/<service>`).
- **Cloud Infrastructure (`deploy/`)**: Designed for zero-cost hosting using Terraform/OpenTofu with Google Cloud Run (scale-to-zero services), GCP Compute Engine `e2-micro` (NATS JetStream), Google Cloud Storage (bird plates and audio), Neon (serverless PostgreSQL), and Cloudflare (DNS and CDN).

---

## Project Status (Work in Progress)

- [x] Polyglot monorepo setup (`go.work`, `uv`, `pnpm`) and GitHub Actions CI/CD pipeline
- [x] Protobuf event & gRPC contracts (`buf`) with generated Go and Python stubs
- [x] Shared Go packages for config (`env`), structured logging (`logger`), and NATS JetStream (`nats`)
- [x] PostgreSQL 18 schema, `sqlc` repository layer, and 435-species Audubon dataset & seed tool
- [x] `catalog` HTTP REST API, health probes, and gRPC crawl job dispatch to `frontier`
- [x] `web` Vue 3 + TypeScript + Tailwind v4 + TanStack Query project scaffolding
- [x] `catalog` NATS consumer for recording sightings and crawl progress
- [ ] `frontier` gRPC server, BFS URL queue, deduplication, and host rate limiting
- [ ] `scraper` HTML fetcher and link/text/image extractor
- [ ] `textparser` Aho-Corasick bird name matcher
- [ ] `vision` Python image classification worker
- [ ] `web` dashboard views for browsing the catalog and monitoring crawls

### Stretch Goals & Extensions

- [ ] **Observability**: OpenTelemetry distributed tracing (`shared/telemetry`) across services
- [ ] **Cloud Hosting**: Terraform / OpenTofu zero-cost infrastructure (`deploy/`)
- [ ] **CLI Client**: Terminal client for dispatching crawls and exploring sightings (`catalog/cmd/cli`)

---

## Data Sources & Attribution

- **Illustrations**: John James Audubon, *The Birds of America* (1827–1838), Public Domain. All bird plates are Courtesy of the John James Audubon Center at Mill Grove, Montgomery County Audubon Collection, and Zebra Publishing.
- **Audio Recordings**: Sourced from [xeno-canto.org](https://xeno-canto.org). Individual clips are licensed under Creative Commons (see `metadata.audio` in `species.json` for specific recordists and licenses).
- **Species Summaries**: Sourced from [Wikipedia](https://wikipedia.org) under CC BY-SA 4.0.

## License

MIT — see [LICENSE](LICENSE).
