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
SELECT g.*, a.password, a.email, a.username, a.access_token FROM public."group" g
JOIN public.account a ON g.account_id = a.id
WHERE g.id = $1;

-- name: UpdateGroupScannedAt :exec
UPDATE public."group"
SET scanned_at = NOW()
WHERE id = $1;

-- name: CreatePost :one
INSERT INTO public.post (post_id, content, created_at, inserted_at, group_id, is_analyzed)
VALUES ($1, $2, $3, NOW(), $4, false)
RETURNING *;

-- name: GetPostById :one
SELECT * FROM public.post WHERE id = $1;

-- name: GetPostByIdWithAccount :one
SELECT p.*, a.password, a.email, a.username, a.access_token, a.id AS account_id FROM public.post p
JOIN public."group" g ON p.group_id = g.id
JOIN public.account a ON g.account_id = a.id
WHERE p.id = $1;

-- name: CreateComment :one
INSERT INTO public.comment (post_id, comment_id, content, created_at, author_id, is_analyzed, inserted_at)
VALUES ($1, $2, $3, $4, $5, false, NOW())
RETURNING *;

-- name: CreateProfile :one
INSERT INTO public.user_profile (facebook_id, name, scraped_by_id, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING *;

-- name: GetProfileById :one
SELECT * FROM public.user_profile WHERE id = $1;

-- name: GetProfileByIdWithAccount :one
SELECT up.*, a.password, a.email, a.username, a.access_token FROM public.user_profile up
JOIN public.account a ON up.scraped_by_id = a.id
WHERE up.id = $1;

-- name: UpdateProfileAfterScan :one
UPDATE public.user_profile
SET updated_at = NOW(),
    is_scanned = TRUE,
    bio = $2,
    location = $3,
    work = $4,
    education = $5,
    relationship_status = $6,
    profile_url = $7,
    hometown = $8,
    locale = $9,
    gender = $10,
    birthday = $11,
    email = $12,
    phone = $13
WHERE id = $1
RETURNING *;

-- name: GetAllConfigs :many
SELECT * FROM public.config;