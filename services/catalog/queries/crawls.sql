-- name: CreateCrawl :one
INSERT    INTO crawls (target_url)
VALUES    ($1)
RETURNING *;

-- name: GetCrawlByID :one
SELECT    *
FROM      crawls
WHERE     id = $1
LIMIT     1;
