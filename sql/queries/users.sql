-- name: CreateUser :one
INSERT INTO
  users (
    id,
    created_at,
    updated_at,
    email,
    hashed_password
  )
VALUES
  ($1, $2, $3, $4, $5)
RETURNING
  *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: UpdateUser :one
UPDATE users
SET
  updated_at = $2,
  email = $3,
  hashed_password = $4
WHERE
  id = $1
RETURNING
  *;

-- name: GetUserByEmail :one
SELECT
  *
FROM
  users
WHERE
  email = $1;

-- name: GetUserFromRefreshToken :one
SELECT
  users.*
FROM
  users
  INNER JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE
  refresh_tokens.token = $1;
