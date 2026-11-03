# Vision Service

The **Vision** service (`services/vision`) is a Python 3.14+ worker managed with [uv](https://docs.astral.sh/uv/). It downloads image URLs extracted during page scraping, runs a computer vision classification model to detect bird species, and publishes sightings to the catalog.

## Responsibilities & Interfaces

- **Image Downloading & Preprocessing**: Fetches and prepares image assets extracted from scraped pages.
- **Species Classification**: Evaluates image contents against bird recognition models mapped to the Audubon species catalog.
- **Event Consumer**:
  - Consumes `PageScraped` (`birdwatcher.v1.page.scraped`) from `scraper`.
- **Event Producer**:
  - Emits `SightingImage` (`birdwatcher.v1.sighting.image`) containing detected species plate number, image bounding data, and model confidence score to `catalog`.

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

Run the service with `uv`:

```bash
uv run vision
```

Run test suite:

```bash
uv run pytest
```

Lint and format:

```bash
uv run ruff check .
uv run ruff format .
```
