-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING *;


-- name: CreateChirp :one
INSERT INTO chirps (chirp_id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING *;


-- name: ListChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;



-- name: GetChirp :one
SELECT * FROM chirps
WHERE chirp_id = $1;


-- name: GetUser :one
SELECT * FROM users
WHERE email = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3
)
RETURNING *;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens SET revoked_at = NOW(),
updated_at = NOW()
WHERE token = $1
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT users.* FROM users
JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1
AND revoked_at IS NULL
AND expires_at > NOW();


-- name: UpdateUser :one
UPDATE users SET email = $1,
hashed_password = $2
WHERE id = $3
RETURNING *;


-- name: DeleteChirp :one
DELETE FROM chirps
WHERE chirp_id = $1
RETURNING *;


-- name: UpgradeChirpyRed :one
UPDATE users SET is_chirpy_red = TRUE
WHERE id = $1
RETURNING *;