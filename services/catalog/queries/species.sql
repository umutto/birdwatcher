-- name: UpsertSpecies :one
INSERT INTO species (plate_number, common_name, scientific_name, description, image_url, audubon_url)
VALUES (@plate_number, @common_name, @scientific_name, @description, @image_url, @audubon_url)
ON CONFLICT (plate_number)
DO UPDATE SET common_name = @common_name,
scientific_name = @scientific_name,
description = @description,
image_url = @image_url,
audubon_url = @audubon_url
RETURNING *;

-- name: GetSpeciesByPlateNumber :one
SELECT *
FROM species
WHERE plate_number = @plate_number;

-- name: ListAllSpecies :many
SELECT *
FROM species
ORDER BY plate_number ASC;

-- name: GetCatalogForCrawl :many
SELECT
    s.*,
    COUNT(st.id) AS sighting_count,
    MAX(st.created_at)::TIMESTAMPTZ AS last_sighted_at
FROM species s
LEFT JOIN sightings st
    ON
        st.species_plate_number = s.plate_number
        AND st.crawl_id = @crawl_id
GROUP BY s.plate_number
ORDER BY s.plate_number;
