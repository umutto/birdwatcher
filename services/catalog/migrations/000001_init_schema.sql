/*  */

-- +goose Up
CREATE TYPE CRAWL_STATE AS ENUM (
    'pending',
    'running',
    'completed',
    'failed',
    'cancelled'
);

CREATE TYPE SIGHTING_SOURCE_TYPE AS ENUM ('text', 'image');

CREATE OR REPLACE FUNCTION SET_UPDATED_AT()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TABLE crawls (
    id UUID PRIMARY KEY DEFAULT UUIDV7(),

    target_url TEXT NOT NULL,
    title TEXT,
    state CRAWL_STATE NOT NULL DEFAULT 'pending',

    pages_crawled_count INT NOT NULL DEFAULT 0,
    images_scanned_count INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
) WITH (fillfactor = 80);

CREATE INDEX idx_crawls_created_at ON crawls (created_at DESC);
CREATE INDEX idx_crawls_state ON crawls (state);

CREATE TRIGGER trigger_crawls_updated_at
BEFORE UPDATE ON crawls
FOR EACH ROW
EXECUTE FUNCTION SET_UPDATED_AT();


CREATE TABLE species (
    plate_number INT PRIMARY KEY,
    common_name TEXT NOT NULL UNIQUE,
    aliases TEXT [] NOT NULL DEFAULT '{}',

    scientific_name TEXT NOT NULL,
    description TEXT,
    image_url TEXT NOT NULL,
    call_url TEXT,
    audubon_url TEXT
);


CREATE TABLE sightings (
    id UUID PRIMARY KEY DEFAULT UUIDV7(),
    crawl_id UUID NOT NULL REFERENCES crawls(id) ON DELETE CASCADE,
    species_plate_number INT NOT NULL REFERENCES species(plate_number) ON DELETE RESTRICT,

    source_type SIGHTING_SOURCE_TYPE NOT NULL,
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


CREATE INDEX idx_sightings_crawl_species ON sightings (crawl_id, species_plate_number);
CREATE INDEX idx_sightings_species_id ON sightings (species_plate_number);

-- +goose Down
DROP TRIGGER IF EXISTS trigger_crawls_updated_at ON crawls;

DROP TABLE IF EXISTS sightings;
DROP TABLE IF EXISTS species;
DROP TABLE IF EXISTS crawls;

DROP TYPE IF EXISTS SIGHTING_SOURCE_TYPE;
DROP TYPE IF EXISTS CRAWL_STATE;

DROP FUNCTION IF EXISTS SET_UPDATED_AT;
