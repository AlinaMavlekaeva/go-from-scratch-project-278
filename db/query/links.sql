-- name: GetLinks :many
SELECT id, 
original_url,
short_name, 
short_url
FROM links
ORDER BY created_at DESC;

-- name: CreateLink :one
INSERT INTO links (original_url, short_name, short_url) 
VALUES ($1, $2, $3) 
RETURNING id, original_url, short_name, short_url;

-- name: GetLinkByID :one
SELECT id, original_url, short_name, short_url
FROM links WHERE id = $1;

-- name: UpdateLinkByID :one
UPDATE links SET original_url = $2,
short_name = $3,
short_url = $4
WHERE id = $1
RETURNING id, original_url, short_name, short_url;

-- name: DeleteLink :exec
DELETE FROM links WHERE id = $1;