-- name: CreateCrawl :one
INSERT INTO crawls (target_url)
VALUES (@target_url)
RETURNING *;

-- name: GetCrawlByID :one
SELECT *
FROM crawls
WHERE id = @id
LIMIT 1;

-- name: UpdateCrawlTitle :exec
UPDATE crawls
SET title = COALESCE(title, @title)
WHERE id = @id;

-- name: UpdateCrawlState :one
UPDATE crawls
SET
    state = @state::CRAWL_STATE,
    started_at = CASE
        WHEN @state::CRAWL_STATE = 'running' THEN COALESCE(started_at, NOW())
        ELSE started_at
    END,
    completed_at = CASE
        WHEN @state::CRAWL_STATE IN ('completed', 'failed', 'cancelled') THEN COALESCE(completed_at, NOW())
        WHEN @state::CRAWL_STATE = 'running' THEN NULL
        ELSE completed_at
    END
WHERE id = @id
RETURNING *;

-- name: IncrementCrawlProgress :exec
UPDATE crawls
SET
    pages_crawled_count = pages_crawled_count + @pages_crawled_count,
    images_scanned_count = images_scanned_count + @images_scanned_count
WHERE id = @id;

-- name: ListCrawlsWithSummary :many
SELECT
    c.*,
    COUNT(s.id) AS total_sightings,
    COUNT(s.id) FILTER (WHERE s.source_type = 'text') AS text_sighting_count,
    COUNT(s.id) FILTER (WHERE s.source_type = 'image') AS image_sighting_count,
    COUNT(DISTINCT s.species_plate_number) AS unique_species_count
FROM crawls AS c
LEFT JOIN sightings AS s
    ON c.id = s.crawl_id
GROUP BY c.id
ORDER BY c.created_at DESC
LIMIT @limit_rows OFFSET @offset_rows;

-- name: DeleteCrawl :execrows
DELETE
FROM crawls
WHERE
    id = @id
    AND state IN ('completed', 'failed', 'cancelled');
