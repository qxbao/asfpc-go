-- name: CreateAccount :one
INSERT INTO public.account (email, username, password, is_block, ua, created_at, updated_at, access_token, proxy_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetAccountById :one
SELECT * FROM public.account WHERE id = $1;

-- name: GetAccounts :many
SELECT * FROM public.account ORDER BY id LIMIT $1 OFFSET $2;

-- name: UpdateAccountAccessToken :one
UPDATE public.account
SET updated_at = NOW(), access_token = $2
WHERE id = $1
RETURNING *;