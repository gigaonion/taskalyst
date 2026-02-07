-- name: CreateApiToken :one
INSERT INTO api_tokens (
    user_id, name, token_hash, expires_at
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: ListApiTokens :many
SELECT id, name, created_at, expires_at 
FROM api_tokens 
WHERE user_id = $1 
ORDER BY created_at DESC;

-- name: DeleteApiToken :exec
DELETE FROM api_tokens 
WHERE id = $1 AND user_id = $2;

-- name: GetUserByTokenHash :one
SELECT u.* FROM api_tokens t
JOIN users u ON t.user_id = u.id
WHERE t.token_hash = $1 
  AND (t.expires_at IS NULL OR t.expires_at > NOW());
