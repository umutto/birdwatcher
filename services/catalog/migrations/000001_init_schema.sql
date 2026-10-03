-- +goose Up
CREATE TYPE crawl_state AS ENUM(
    'pending',
    'running',
    'completed',
    'failed',
    'cancelled'
);

CREATE TYPE sighting_source_type AS ENUM('text', 'image');

CREATE TABLE crawls (
    id UUID PRIMARY KEY DEFAULT UUIDV7(),
    target_url TEXT NOT NULL,
    title TEXT,
    state crawl_state NOT NULL DEFAULT 'pending',
    pages_crawled_count INT NOT NULL DEFAULT 0,
    images_scanned_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
)
WITH
    (fillfactor = 80);

CREATE INDEX idx_crawls_created_at ON crawls (created_at DESC);

CREATE INDEX idx_crawls_state ON crawls (state);

CREATE TABLE species (
    id UUID PRIMARY KEY DEFAULT UUIDV7(),
    plate_number INT NOT NULL UNIQUE,
    common_name TEXT NOT NULL UNIQUE,
    scientific_name TEXT NOT NULL,
    description TEXT,
    image_url TEXT NOT NULL,
    audubon_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sightings (
    id UUID PRIMARY KEY DEFAULT UUIDV7(),
    crawl_id UUID NOT NULL REFERENCES crawls (id) ON DELETE CASCADE,
    species_id UUID NOT NULL REFERENCES species (id) ON DELETE RESTRICT,
    source_type sighting_source_type NOT NULL,
    confidence REAL NOT NULL DEFAULT 0.0,
    page_url TEXT NOT NULL,
    snippet TEXT,
    image_url TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT check_sighting_source_type_payload CHECK (
        (
            source_type = 'text'
            AND snippet IS NOT NULL
        )
        OR (
            source_type = 'image'
            AND image_url IS NOT NULL
        )
    )
);

CREATE INDEX idx_sightings_crawl_species ON sightings (crawl_id, species_id);

CREATE INDEX idx_sightings_species_id ON sightings (species_id);

-- +goose Down
DROP TABLE IF EXISTS sightings;

DROP TABLE IF EXISTS species;

DROP TABLE IF EXISTS crawls;

DROP TYPE if EXISTS sighting_source_type;

DROP TYPE if EXISTS crawl_state;
