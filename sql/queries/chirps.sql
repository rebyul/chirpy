-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
    VALUES (gen_random_uuid (), now(), now(), $1, $2)
RETURNING
    *;

-- name: GetChirps :many
SELECT
    id,
    created_at,
    updated_at,
    body,
    user_id
FROM
    chirps
WHERE
    sqlc.narg ('author_filter')::uuid IS NULL
    OR user_id = sqlc.narg ('author_filter')::uuid
ORDER BY
    created_at ASC;

-- name: GetChirpById :one
SELECT
    id,
    created_at,
    updated_at,
    body,
    user_id
FROM
    chirps
WHERE
    id = $1;

-- name: DeleteChirpById :exec
DELETE FROM chirps
WHERE id = $1;

