# Textparser Service

The **Textparser** service (`services/textparser`) scans scraped web page text to detect bird species. It uses Aho-Corasick automaton matching and text normalization against the 435 Audubon species dictionary (including scientific names and common aliases).

## Responsibilities & Interfaces

- **Pattern Matching**: Implements linear-time multi-pattern string matching over page content to find species occurrences without repeated regex scans.
- **Event Consumer**:
  - Consumes `PageScraped` (`birdwatcher.v1.page.scraped`) from `scraper`.
- **Event Producer**:
  - Emits `SightingText` (`birdwatcher.v1.sighting.text`) containing the detected species plate number, match context snippet, and confidence score to `catalog`.

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
