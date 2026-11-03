# Catalog Service

The **Catalog** service (`services/catalog`) is the primary data management and query gateway of the birdwatcher pipeline. It owns the PostgreSQL database, ingests sightings and crawl events from NATS JetStream, dispatches crawl jobs to `frontier` via gRPC, and exposes an HTTP REST API for the web dashboard (`web`) and CLI clients (`cmd/cli`).

## Architecture & Responsibilities

- **Storage & Migrations**: Manages species, sightings, and crawl metadata using PostgreSQL 18 with [sqlc](https://sqlc.dev/) for type-safe SQL queries and [goose](https://github.com/pressly/goose) for schema migrations.
- **REST API**: Serves endpoints on port `8080` for managing crawls and viewing the Audubon bird catalog.
- **Event Consumer**: Listens to the `BIRDWATCHER` stream on NATS JetStream for `PageScraped`, `SightingText`, `SightingImage`, and `JobCompleted` events to update progress and persist observations.
- **Job Dispatch**: Connects via gRPC to `frontier` (`FrontierService.StartJob`) when a crawl is initiated.

---

## HTTP REST API

The HTTP server listens on port `8080`:

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

## Configuration

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

| Variable | Default | Description |
| :--- | :--- | :--- |
| `ENV` | `development` | Environment mode (`development`, `production`). |
| `PORT` | `8080` | HTTP server port. |
| `FRONTIER_GRPC_ADDR` | `localhost:9090` | Address of the Frontier gRPC server. |
| `FRONTIER_MAX_DEPTH` | `4` | Default crawl depth limit. |
| `FRONTIER_MAX_PAGES` | `200` | Default maximum pages per crawl. |
| `NATS_URL` | `nats://localhost:4222` | NATS JetStream connection URL. |
| `DATABASE_URL` | `postgres://birdwatcher:birdwatcher@localhost:5432/birdwatcher?sslmode=disable` | PostgreSQL connection string. |

---

## Development

Run with live reload from the repository root:

```bash
make dev
```

Or run directly:

```bash
go run ./cmd/catalog
```
