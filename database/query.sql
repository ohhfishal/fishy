-- name: GetLastJob :many
SELECT * FROM jobs
ORDER BY created_at DESC
LIMIT 1;

-- name: PutJob :one
INSERT INTO jobs (
  failures
) values (?) 
RETURNING *;

-- name: GetCards :many
SELECT * FROM flashcards;

-- name: InsertCard :one
INSERT INTO flashcards (
  header,
  description,
  origin,
  class_context,
  ai_overview,
  thumbnail
) values (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: Metrics :one
SELECT
  (SELECT COUNT(*) FROM jobs) as jobs,
  (SELECT COUNT(*) FROM flashcards) as flashcards;
