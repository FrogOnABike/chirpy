-- name: CreateRToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
   $1,
    NOW(),
    NOW(),
    $2,
    NOW() + INTERVAL '60 days',
    NULL 
)
RETURNING *;

-- name: GetUserFromRToken :one
SELECT refresh_tokens.user_id
FROM refresh_tokens
WHERE refresh_tokens.token = $1 AND refresh_tokens.expires_at > NOW() AND refresh_tokens.revoked_at IS NULL;

