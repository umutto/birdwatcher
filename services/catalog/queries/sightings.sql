-- name: CreateSighting :one
INSERT INTO sightings (crawl_id, species_plate_number, source_type, confidence, page_url, snippet, image_url)
VALUES (@crawl_id, @species_plate_number, @source_type, @confidence, @page_url, @snippet, @image_url)
RETURNING *;

-- name: ListSightingsByCrawl :many
SELECT
    st.*,
    s.common_name,
    s.scientific_name
FROM sightings st
INNER JOIN species s
    ON st.species_plate_number = s.plate_number
WHERE st.crawl_id = @crawl_id
ORDER BY st.created_at DESC
LIMIT @limit_rows OFFSET @offset_rows;

-- name: ListSightingsBySpecies :many
SELECT *
FROM sightings
WHERE species_plate_number = @species_plate_number
ORDER BY created_at DESC
LIMIT @limit_rows OFFSET @offset_rows;
