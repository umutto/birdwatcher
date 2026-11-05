# Frontier Service

The **Frontier** service (`services/frontier`) coordinates web crawling traversal. It exposes a gRPC interface to accept new crawl jobs, manages the URL queue using a breadth-first search (BFS) strategy, prevents duplicate page visits with Valkey, enforces politeness, and applies per-host rate limiting.

## Responsibilities & Interfaces

- **gRPC API (`:9090`)**: Implements `FrontierService.StartJob` (defined in `proto/birdwatcher/v1/frontier.proto`) called by the `catalog` service.
- **Event Producer**: Publishes `PageFetchRequested` (`birdwatcher.v1.page.fetch`) to NATS JetStream for scraper workers to fetch, and `JobCompleted` (`birdwatcher.v1.job.completed`) when a crawl terminates.
- **Event Consumer**: Listens for `UrlDiscovered` (`birdwatcher.v1.url.discovered`) published by `scraper` to schedule new candidate URLs within crawl depth limits.

---

## Architecture & Scalability

- **Deduplication**: Backed by Valkey Set (`visited:<jobID>`) for durable, cross-restart state.
- **Queue & Politeness**: Kept in-memory for zero-latency scheduling while workers scale horizontally via NATS.
- **Horizontal Scaling Note**: Scaling Frontier itself to multiple nodes would require **domain sharding with consistent hashing** to preserve politeness without distributed lock contention.

---

## Configuration

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

| Variable | Default | Description |
| :--- | :--- | :--- |
| `ENV` | `development` | Environment mode (`development`, `production`). |
| `PORT` | `9090` | gRPC server port. |
| `NATS_URL` | `nats://localhost:4222` | NATS JetStream connection URL. |
| `VALKEY_ADDR` | `localhost:6379` | Valkey/Redis instance address for visited URL deduplication. |
| `CRAWL_HOST_DELAY` | `250ms` | Minimum delay between subsequent requests to the same host. |
| `CRAWL_MAX_INFLIGHT_PER_HOST` | `2` | Maximum concurrent HTTP fetches allowed per host. |

---

## Development

Run the service directly:

```bash
go run .
```

Run unit and integration tests:

```bash
go test ./...
```
