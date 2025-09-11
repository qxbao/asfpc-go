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

-- name: CreateGroup :one
INSERT INTO public."group" (group_id, group_name, is_joined, account_id)
VALUES ($1, $2, false, $3)
RETURNING *;

-- name: GetGroupById :one
SELECT * FROM public."group" WHERE id = $1;

-- name: GetGroupByIdWithAccount :one
SELECT g.*, a.* FROM public."group" g
JOIN public.account a ON g.account_id = a.id
WHERE g.id = $1;

-- name: CreatePost :one
INSERT INTO public.post (post_id, content, created_at, inserted_at, group_id, is_analyzed)
VALUES ($1, $2, $3, NOW(), $4, false)
RETURNING *;

