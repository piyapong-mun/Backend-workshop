-- name: CreateSessions :one
INSERT INTO sessions (
    ID,
    username, 
    refresh_token, 
    user_agent, 
    client_ip, 
    is_blocked, 
    expired_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetSessions :one
SELECT * FROM sessions
WHERE ID = $1 LIMIT 1; 