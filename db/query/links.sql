-- name: GetLinks :many
SELECT id, 
original_url,
short_name, 
short_url
FROM links 
ORDER BY created_at DESC;

-- name: GetLinksPage :many
SELECT id, 
original_url,
short_name, 
short_url
FROM links 
ORDER BY created_at DESC LIMIT $2 OFFSET $1;

-- name: CountLinks :one
SELECT COUNT (*) FROM links;

-- name: CreateLink :one
INSERT INTO links (original_url, short_name, short_url) 
VALUES ($1, $2, $3) 
RETURNING id, original_url, short_name, short_url;

-- name: GetLinkByID :one
SELECT id, original_url, short_name, short_url
FROM links WHERE id = $1;

-- name: GetLinkByShortName :one
SELECT id, original_url, short_name, short_url
FROM links WHERE short_name = $1;

-- name: UpdateLinkByID :one
UPDATE links SET original_url = $2,
short_name = $3,
short_url = $4
WHERE id = $1
RETURNING id, original_url, short_name, short_url;

-- name: DeleteLink :exec
DELETE FROM links WHERE id = $1;

-- name: CreateLinkVisit :one
INSERT INTO link_visits (link_id,
ip,
user_agent,
referer,
status) 
VALUES ($1, $2, $3, $4, $5) 
RETURNING id, link_id, ip, user_agent, status;

-- name: GetLinkVisits :many
SELECT id, 
link_id,
created_at, 
ip,
user_agent,
status
FROM link_visits
ORDER BY created_at DESC;

-- name: GetLinkVisitsPage :many
SELECT id, 
link_id,
created_at, 
ip,
user_agent,
status
FROM link_visits 
ORDER BY created_at DESC LIMIT $2 OFFSET $1;

-- name: CountLinkVisits :one
SELECT COUNT (*) FROM link_visits;