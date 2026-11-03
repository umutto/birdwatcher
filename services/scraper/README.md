# Scraper Service

The **Scraper** service (`services/scraper`) is a stateless, horizontally scalable worker that fetches HTML web pages, parses their DOM, extracts outgoing links (`<a>`), and collects page text and image URLs (`<img>`) for downstream bird detection.

## Responsibilities & Interfaces

- **Page Fetching**: Pulls fetch tasks asynchronously from NATS JetStream and retrieves pages with proper timeouts, user-agent headers, and error handling.
- **HTML Extraction**: Parses HTML using `net/http` and `goquery` to extract page titles, main body text, and links.
- **Event Consumer**:
  - Consumes `PageFetchRequested` (`birdwatcher.v1.page.fetch`) from `frontier`.
- **Event Producer**:
  - Emits `UrlDiscovered` (`birdwatcher.v1.url.discovered`) back to `frontier` for URL queueing and deduplication.
  - Emits `PageScraped` (`birdwatcher.v1.page.scraped`) to notify `textparser`, `vision`, and `catalog`.

---

## Configuration

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

| Variable | Default | Description |
| :--- | :--- | :--- |
| `ENV` | `development` | Environment mode (`development`, `production`). |
| `NATS_URL` | `nats://localhost:4222` | NATS JetStream connection URL. |

---

## Development

Run the service directly:

```bash
go run .
```

Run tests:

```bash
go test ./...
```
