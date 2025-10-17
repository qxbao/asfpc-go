-- name: CreateAccount :one
INSERT INTO public.account (email, username, password, is_block, ua, created_at, updated_at, access_token, proxy_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetAccountById :one
SELECT * FROM public.account WHERE id = $1;

-- name: GetAccountStats :one
SELECT
  (SELECT COUNT(*) FROM public.account) AS total_accounts,
  (SELECT COUNT(*) FROM public.account WHERE is_block = false and access_token IS NOT NULL) AS active_accounts,
  (SELECT COUNT(*) FROM public.account WHERE is_block = true) AS blocked_accounts;

-- name: GetAccounts :many
SELECT a.id, a.username, a.email, a.updated_at, a.access_token, (
	SELECT COUNT(*) FROM public."group" g WHERE g.account_id = a.id
) as group_count, COOKIES IS NOT NULL as is_login,
a.is_block
FROM public.account a LIMIT $1 OFFSET $2;

-- name: GetOKAccountIds :many
SELECT t.id
FROM (SELECT a.id,
  (SELECT COUNT(*) FROM public."group" g WHERE g.account_id = a.id) AS group_count
  FROM public.account a
  WHERE a.is_block = false AND a.access_token IS NOT NULL
) t
WHERE t.group_count > 0;

-- name: UpdateAccountAccessToken :one
UPDATE public.account
SET updated_at = NOW(), access_token = $2
WHERE id = $1
RETURNING *;

-- name: UpdateAccountCredentials :one
UPDATE public.account
SET updated_at = NOW(),
    email = $2,
    username = $3,
    password = $4
WHERE id = $1
RETURNING *;

-- name: DeleteAccounts :exec
DELETE FROM public.account WHERE id = ANY($1::int[]);

-- name: CreateGroup :one
INSERT INTO public."group" (group_id, group_name, is_joined, account_id)
VALUES ($1, $2, false, $3)
RETURNING *;

-- name: GetGroupCategories :many
SELECT c.* FROM public.category c
JOIN public.group_category gc ON c.id = gc.category_id
WHERE gc.group_id = $1;

-- name: AddGroupCategory :exec
INSERT INTO public.group_category (group_id, category_id)
VALUES ($1, $2)
ON CONFLICT (group_id, category_id) DO NOTHING;

-- name: DeleteGroupCategory :exec
DELETE FROM public.group_category WHERE group_id = $1 AND category_id = $2;

-- name: DeleteGroup :exec
WITH deleted_posts AS (
  DELETE FROM public.post WHERE group_id = $1 RETURNING post.id
),
deleted_comments AS (
  DELETE FROM public.comment WHERE post_id IN (SELECT id FROM deleted_posts)
)
DELETE FROM public."group" WHERE "group".id = $1;

-- name: GetGroupsByAccountId :many
SELECT *, COALESCE((
  SELECT json_agg(c) FROM public.category c
  JOIN public.group_category gc ON c.id = gc.category_id
  WHERE gc.group_id = g.id
), '[]'::json)::jsonb as categories FROM public."group" g WHERE account_id = $1;

-- name: GetGroupsToScan :many
SELECT g.*, a.access_token, COALESCE((
  SELECT json_agg(c) FROM public.category c
  JOIN public.group_category gc ON c.id = gc.category_id
  WHERE gc.group_id = g.id
), '[]'::json)::jsonb as categories FROM public."group" g
JOIN public.account a ON g.account_id = a.id
WHERE g.is_joined = true AND g.account_id = $1
ORDER BY scanned_at ASC NULLS FIRST LIMIT $2;

-- name: UpdateGroupScannedAt :exec
UPDATE public."group"
SET scanned_at = NOW()
WHERE id = $1;

-- name: CreatePost :one
INSERT INTO public.post (post_id, content, created_at, inserted_at, group_id, is_analyzed)
VALUES ($1, $2, $3, NOW(), $4, true)
ON CONFLICT (post_id) DO UPDATE SET
    id = EXCLUDED.id
RETURNING *;

-- name: CreateComment :one
INSERT INTO public.comment (post_id, comment_id, content, created_at, author_id, is_analyzed, inserted_at)
VALUES ($1, $2, $3, $4, $5, false, NOW())
ON CONFLICT (comment_id) DO UPDATE SET
    id = EXCLUDED.id
RETURNING *;

-- name: CreateProfile :one
INSERT INTO public.user_profile (facebook_id, name, scraped_by_id, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (facebook_id) DO UPDATE SET
    id = EXCLUDED.id
RETURNING *;

-- name: AddUserProfileCategory :exec
INSERT INTO public.user_profile_category (user_profile_id, category_id)
VALUES ($1, $2)
ON CONFLICT (user_profile_id, category_id) DO NOTHING;

-- name: GetProfileById :one
SELECT *, COALESCE((SELECT json_agg(c) FROM public.user_profile_category c WHERE c.user_profile_id = up.id), '[]'::json)::jsonb as categories
FROM public.user_profile up WHERE id = $1;

-- name: GetProfilesToScan :many
SELECT up.*, a.access_token, a.id as account_id
FROM public.user_profile up
JOIN public.account a ON up.scraped_by_id = a.id
WHERE up.is_scanned = false AND a.is_block = false AND a.access_token IS NOT NULL
ORDER BY up.updated_at ASC LIMIT $1;

-- name: GetProfilesAnalysisPage :many
SELECT 
  up.id,
  up.facebook_id,
  up.name,
  up.is_analyzed,
  up.gemini_score,
  up.model_score,
  (COALESCE((
    SELECT json_agg(json_build_object('id', c.id, 'name', c.name, 'description', c.description))
    FROM public.user_profile_category upc
    JOIN public.category c ON c.id = upc.category_id
    WHERE upc.user_profile_id = up.id
  ), '[]'::json))::jsonb as categories,
  ((COALESCE(up.bio, '') != '')::int +
  (COALESCE(up.location, '') != '')::int +
  (COALESCE(up.work, '') != '')::int +
  (COALESCE(up.locale, '') != '')::int +
  (COALESCE(up.education, '') != '')::int +
  (COALESCE(up.relationship_status, '') != '')::int +
  (COALESCE(up.hometown, '') != '')::int +
  (COALESCE(up.gender, '') != '')::int +
  (COALESCE(up.birthday, '') != '')::int +
  (COALESCE(up.email, '') != '')::int +
  (COALESCE(up.phone, '') != '')::int)::int AS non_null_count
FROM public.user_profile up
WHERE up.is_scanned = true
ORDER BY up.model_score DESC NULLS LAST, up.gemini_score DESC NULLS LAST, non_null_count DESC, up.updated_at ASC
LIMIT $1 OFFSET $2;

-- name: GetProfilesAnalysisPageByCategory :many
SELECT 
  up.id,
  up.facebook_id,
  up.name,
  up.is_analyzed,
  up.gemini_score,
  up.model_score,
  (COALESCE((
    SELECT json_agg(json_build_object('id', c.id, 'name', c.name, 'description', c.description))
    FROM public.user_profile_category upc
    JOIN public.category c ON c.id = upc.category_id
    WHERE upc.user_profile_id = up.id
  ), '[]'::json))::jsonb as categories,
  ((COALESCE(up.bio, '') != '')::int +
  (COALESCE(up.location, '') != '')::int +
  (COALESCE(up.work, '') != '')::int +
  (COALESCE(up.locale, '') != '')::int +
  (COALESCE(up.education, '') != '')::int +
  (COALESCE(up.relationship_status, '') != '')::int +
  (COALESCE(up.hometown, '') != '')::int +
  (COALESCE(up.gender, '') != '')::int +
  (COALESCE(up.birthday, '') != '')::int +
  (COALESCE(up.email, '') != '')::int +
  (COALESCE(up.phone, '') != '')::int)::int AS non_null_count
FROM public.user_profile up
JOIN public.user_profile_category upc ON up.id = upc.user_profile_id
WHERE up.is_scanned = true AND upc.category_id = $1
ORDER BY up.model_score DESC NULLS LAST, up.gemini_score DESC NULLS LAST, non_null_count DESC, up.updated_at ASC
LIMIT $2 OFFSET $3;

-- name: GetProfilesAnalysisCronjob :many
SELECT up.*,
  upc.category_id,
  (COALESCE(up.bio, '') != '')::int +
  (COALESCE(up.location, '') != '')::int +
  (COALESCE(up.work, '') != '')::int +
  (COALESCE(up.locale, '') != '')::int +
  (COALESCE(up.education, '') != '')::int +
  (COALESCE(up.relationship_status, '') != '')::int +
  (COALESCE(up.hometown, '') != '')::int +
  (COALESCE(up.gender, '') != '')::int +
  (COALESCE(up.birthday, '') != '')::int +
  (COALESCE(up.email, '') != '')::int +
  (COALESCE(up.phone, '') != '')::int AS non_null_count
FROM public.user_profile up
JOIN public.user_profile_category upc ON up.id = upc.user_profile_id
WHERE up.is_scanned = true AND up.is_analyzed = false
AND upc.category_id = $1
ORDER BY non_null_count DESC, up.updated_at ASC
LIMIT $2;

-- name: GetProfileStats :one
SELECT
  (SELECT COUNT(*) FROM public.user_profile) AS total_profiles,
  (SELECT COUNT(*) FROM public.embedded_profile) AS embedded_count,
  (SELECT COUNT(*) FROM public.user_profile WHERE is_scanned = true) AS scanned_profiles,
  (SELECT COUNT(*) FROM public.user_profile WHERE model_score IS NOT NULL) AS scored_profiles,
  (SELECT COUNT(*) FROM public.user_profile WHERE is_analyzed = true) AS analyzed_profiles;

-- name: GetProfileIDForEmbedding :many
SELECT up.id FROM public.user_profile up
JOIN public.user_profile_category upc ON up.id = upc.user_profile_id
WHERE up.id NOT IN (
  SELECT pid FROM public.embedded_profile
) AND up.is_scanned = true 
AND upc.category_id = $1
LIMIT $2;

-- name: UpsertEmbeddedProfiles :exec
INSERT INTO public.embedded_profile (pid, embedding, created_at)
VALUES ($1, $2, NOW())
ON CONFLICT (pid) DO UPDATE SET
    embedding = EXCLUDED.embedding,
    created_at = NOW();

-- name: CountProfiles :one
SELECT COUNT(*) as total_profiles FROM public.user_profile WHERE is_scanned = true;

-- name: UpdateProfileScanStatus :one
UPDATE public.user_profile
SET updated_at = NOW(),
    is_scanned = TRUE
WHERE id = $1
RETURNING *;

-- name: ResetProfilesModelScore :exec
UPDATE public.user_profile
SET model_score = NULL;

-- name: UpdateGeminiAnalysisProfile :one
UPDATE public.user_profile
SET gemini_score = $2,
    is_analyzed = TRUE,
    updated_at = NOW()
WHERE id = $1
RETURNING gemini_score;

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

-- name: GetProfilesForExport :many
SELECT up.*, ep.embedding FROM public.user_profile up
JOIN public.embedded_profile ep ON up.id = ep.pid;

-- name: GetProfilesForScoring :many
SELECT up.id FROM public.user_profile up
JOIN public.embedded_profile ep ON up.id = ep.pid
JOIN public.user_profile_category upc ON up.id = upc.user_profile_id
WHERE up.is_scanned = true AND up.model_score IS NULL
AND upc.category_id = $1
LIMIT $2;

-- name: UpdateModelScore :exec
UPDATE public.user_profile
SET model_score = $2
WHERE id = $1;

-- name: ImportProfile :one
INSERT INTO public.user_profile (facebook_id, name, bio, location, work, education, relationship_status, created_at, updated_at, scraped_by_id, is_scanned, hometown, locale, gender, birthday, email, phone, profile_url, is_analyzed, gemini_score)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 1, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
ON CONFLICT (facebook_id) DO UPDATE SET
    name = EXCLUDED.name,
    bio = EXCLUDED.bio,
    location = EXCLUDED.location,
    work = EXCLUDED.work,
    education = EXCLUDED.education,
    relationship_status = EXCLUDED.relationship_status,
    updated_at = EXCLUDED.updated_at,
    is_scanned = EXCLUDED.is_scanned,
    hometown = EXCLUDED.hometown,
    locale = EXCLUDED.locale,
    gender = EXCLUDED.gender,
    birthday = EXCLUDED.birthday,
    email = EXCLUDED.email,
    phone = EXCLUDED.phone,
    profile_url = EXCLUDED.profile_url,
    is_analyzed = EXCLUDED.is_analyzed,
    gemini_score = EXCLUDED.gemini_score
RETURNING *;

-- name: DeleteJunkProfiles :one
WITH non_null_count AS (
  SELECT up.id,
    (COALESCE(up.bio, '') != '')::int +
    (COALESCE(up.location, '') != '')::int +
    (COALESCE(up.work, '') != '')::int +
    (COALESCE(up.locale, '') != '')::int +
    (COALESCE(up.education, '') != '')::int +
    (COALESCE(up.relationship_status, '') != '')::int +
    (COALESCE(up.hometown, '') != '')::int +
    (COALESCE(up.gender, '') != '')::int +
    (COALESCE(up.birthday, '') != '')::int +
    (COALESCE(up.email, '') != '')::int +
    (COALESCE(up.phone, '') != '')::int AS field_count
  FROM public.user_profile up
),
profiles_to_delete AS (
  SELECT nnc.id 
  FROM non_null_count nnc
  JOIN public.user_profile up ON nnc.id = up.id
  WHERE
    up.is_scanned = true
    AND (up.name = ''
    OR up.name IS NULL
    OR up.name LIKE '%Anonymous%' 
    OR up.name LIKE '%anonymous%' 
    OR up.name LIKE '%ẩn danh%'
    OR up.name LIKE '%Ẩn danh%'
    OR nnc.field_count < 1)
),
deleted_comments AS (
  DELETE FROM public.comment 
  WHERE author_id IN (SELECT id FROM profiles_to_delete)
  RETURNING author_id
),
deleted_profiles AS (
  DELETE FROM public.user_profile 
  WHERE id IN (SELECT id FROM profiles_to_delete)
  RETURNING id
)
SELECT COUNT(*) as deleted_count FROM deleted_profiles;

-- name: GetPrompt :one
SELECT * FROM public.prompt
WHERE service_name = $1 AND category_id = $2
ORDER BY version DESC LIMIT 1;

-- name: GetAllPrompts :many
SELECT *
FROM (
  SELECT *, ROW_NUMBER() OVER (PARTITION BY service_name ORDER BY version DESC) AS rn
  FROM public.prompt
) t
WHERE rn = 1
ORDER BY service_name
LIMIT $1 OFFSET $2;

-- name: GetPromptsByCategory :many
SELECT *
FROM (
  SELECT *, ROW_NUMBER() OVER (PARTITION BY service_name ORDER BY version DESC) AS rn
  FROM public.prompt WHERE category_id = $3
) t
WHERE rn = 1
ORDER BY service_name
LIMIT $1 OFFSET $2;

-- name: CountPrompts :one
SELECT COUNT(DISTINCT service_name) as total_prompt FROM public.prompt;

-- name: CreatePrompt :one
WITH next_version AS (
  SELECT COALESCE(MAX(version), 0) + 1 AS version
  FROM public.prompt
  WHERE service_name = $1 AND category_id = $4
)
INSERT INTO public.prompt (service_name, version, content, created_by, created_at, category_id)
SELECT $1, next_version.version, $2, $3, NOW(), $4
FROM next_version
RETURNING *;

-- name: DeletePrompt :exec
WITH prompt_to_delete AS (
  SELECT * FROM public.prompt d WHERE d.id = $1
)
DELETE FROM public.prompt WHERE service_name = (SELECT service_name FROM prompt_to_delete) AND category_id = (SELECT category_id FROM prompt_to_delete);

-- name: RollbackPrompt :exec 
DELETE FROM public.prompt pr WHERE
pr.service_name = $1 AND pr.category_id = $2
AND pr.version = (
  SELECT MAX(version) FROM public.prompt p
  WHERE p.service_name = $1 AND p.category_id = $2
);

-- name: GetAllConfigs :many
SELECT * FROM public.config;

-- name: GetConfigByKey :one
SELECT * FROM public.config WHERE "key" = $1;

-- name: UpsertConfig :one
INSERT INTO public.config ("key", "value")
VALUES ($1, $2)
ON CONFLICT ("key") DO UPDATE SET "value" = $2
RETURNING *;

-- name: GetStats :one
SELECT
  (SELECT COUNT(*) FROM public."group") AS total_groups,
  (SELECT COUNT(*) FROM public.comment) AS total_comments,
  (SELECT COUNT(*) FROM public.post) AS total_posts;

-- name: LogAction :exec
INSERT INTO public.log (account_id, "action", target_id, description, created_at)
VALUES ($1, $2, $3, $4, NOW());

-- name: GetLogs :many
SELECT l.*, a.username FROM public.log l
LEFT JOIN public.account a ON l.account_id = a.id
ORDER BY l.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountLogs :one
SELECT COUNT(*) as total_logs FROM public.log;

-- name: GetGeminiKeys :many
SELECT * FROM public.gemini_key;

-- name: GetGeminiKeyForUse :one
SELECT * FROM public.gemini_key ORDER BY updated_at ASC NULLS FIRST LIMIT 1;

-- name: CountGeminiKeys :one
SELECT COUNT(*) as total_gemini_keys FROM public.gemini_key;

-- name: CreateGeminiKey :one
INSERT INTO public.gemini_key (api_key)
VALUES ($1)
RETURNING *;

-- name: DeleteGeminiKey :exec
DELETE FROM public.gemini_key WHERE id = $1;

-- name: UpdateGeminiKeyUsage :one
UPDATE public.gemini_key
SET token_used = token_used + $2,
updated_at = NOW()
WHERE api_key = $1
RETURNING *;

-- name: CreateRequest :one
INSERT INTO public.request(description)
VALUES ($1)
RETURNING id;

-- name: UpdateRequestStatus :exec
UPDATE public.request
SET status = $2, updated_at = NOW(), error_message = $3, progress = $4, description = $5
WHERE id = $1;

-- name: GetRequestById :one
SELECT * FROM public.request WHERE id = $1;

-- name: GetProfileEmbedding :one
SELECT embedding FROM public.embedded_profile WHERE pid = $1;

-- name: FindSimilarProfiles :many
SELECT
  p.id AS profile_id,
  p.profile_url as profile_url,
  p.name AS profile_name,
  CAST(1 - (ep.embedding <=> (
	SELECT embedding FROM public.embedded_profile WHERE embedded_profile.pid = $1
  )) AS DOUBLE PRECISION) AS similarity
FROM embedded_profile ep
JOIN user_profile p ON p.id = ep.pid
WHERE ep.pid != $1
ORDER BY ep.embedding <=> (
	SELECT embedding FROM public.embedded_profile WHERE embedded_profile.pid = $1
  )
LIMIT $2;

-- name: GetDashboardStats :one
SELECT
  (SELECT COUNT(*) FROM public."group") AS total_groups,
  (SELECT COUNT(*) FROM public.comment) AS total_comments,
  (SELECT COUNT(*) FROM public.post) AS total_posts,
  (SELECT COUNT(*) FROM public.user_profile) AS total_profiles,
  (SELECT COUNT(*) FROM public.embedded_profile) AS embedded_count,
  (SELECT COUNT(*) FROM public.user_profile WHERE is_scanned = true) AS scanned_profiles,
  (SELECT COUNT(*) FROM public.user_profile WHERE model_score IS NOT NULL) AS scored_profiles,
  (SELECT COUNT(*) FROM public.user_profile WHERE is_analyzed = true) AS analyzed_profiles,
  (SELECT COUNT(*) FROM public.account) AS total_accounts,
  (SELECT COUNT(*) FROM public.account WHERE is_block = false and access_token IS NOT NULL) AS active_accounts,
  (SELECT COUNT(*) FROM public.account WHERE is_block = true) AS blocked_accounts;

-- name: GetTimeSeriesData :many
SELECT 
  DATE_TRUNC('day', updated_at)::date as date,
  COUNT(*) as count
FROM public.user_profile 
WHERE updated_at >= NOW() - INTERVAL '6 months'
GROUP BY DATE_TRUNC('day', updated_at)
ORDER BY date;

-- name: GetTimeSeriesDataByCategory :many
SELECT 
  DATE_TRUNC('day', up.updated_at)::date as date,
  COUNT(*) as count
FROM public.user_profile up
JOIN public.user_profile_category upc ON up.id = upc.user_profile_id
WHERE up.updated_at >= NOW() - INTERVAL '6 months'
  AND upc.category_id = $1
GROUP BY DATE_TRUNC('day', up.updated_at)
ORDER BY date;

-- name: GetScoreDistribution :many
WITH score_ranges AS (
  SELECT '0.0-0.2' as range UNION ALL
  SELECT '0.2-0.4' UNION ALL
  SELECT '0.4-0.6' UNION ALL
  SELECT '0.6-0.8' UNION ALL
  SELECT '0.8-1.0'
),
gemini_counts AS (
  SELECT
    CASE 
      WHEN gemini_score BETWEEN 0.0 AND 0.2 THEN '0.0-0.2'
      WHEN gemini_score BETWEEN 0.2 AND 0.4 THEN '0.2-0.4'
      WHEN gemini_score BETWEEN 0.4 AND 0.6 THEN '0.4-0.6'
      WHEN gemini_score BETWEEN 0.6 AND 0.8 THEN '0.6-0.8'
      WHEN gemini_score BETWEEN 0.8 AND 1.0 THEN '0.8-1.0'
    END as score_range,
    COUNT(*) as gemini_count
  FROM public.user_profile
  WHERE gemini_score IS NOT NULL
  GROUP BY score_range
),
model_counts AS (
  SELECT
    CASE 
      WHEN model_score BETWEEN 0.0 AND 0.2 THEN '0.0-0.2'
      WHEN model_score BETWEEN 0.2 AND 0.4 THEN '0.2-0.4'
      WHEN model_score BETWEEN 0.4 AND 0.6 THEN '0.4-0.6'
      WHEN model_score BETWEEN 0.6 AND 0.8 THEN '0.6-0.8'
      WHEN model_score BETWEEN 0.8 AND 1.0 THEN '0.8-1.0'
    END as score_range,
    COUNT(*) as model_count
  FROM public.user_profile
  WHERE model_score IS NOT NULL
  GROUP BY score_range
)
SELECT 
  sr.range as score_range,
  COALESCE(gc.gemini_count, 0) as gemini_count,
  COALESCE(mc.model_count, 0) as model_count
FROM score_ranges sr
LEFT JOIN gemini_counts gc ON sr.range = gc.score_range
LEFT JOIN model_counts mc ON sr.range = mc.score_range
ORDER BY sr.range;

-- name: GetScoreDistributionByCategory :many
WITH score_ranges AS (
  SELECT '0.0-0.2' as range UNION ALL
  SELECT '0.2-0.4' UNION ALL
  SELECT '0.4-0.6' UNION ALL
  SELECT '0.6-0.8' UNION ALL
  SELECT '0.8-1.0'
),
gemini_counts AS (
  SELECT
    CASE 
      WHEN up.gemini_score BETWEEN 0.0 AND 0.2 THEN '0.0-0.2'
      WHEN up.gemini_score BETWEEN 0.2 AND 0.4 THEN '0.2-0.4'
      WHEN up.gemini_score BETWEEN 0.4 AND 0.6 THEN '0.4-0.6'
      WHEN up.gemini_score BETWEEN 0.6 AND 0.8 THEN '0.6-0.8'
      WHEN up.gemini_score BETWEEN 0.8 AND 1.0 THEN '0.8-1.0'
    END as score_range,
    COUNT(*) as gemini_count
  FROM public.user_profile up
  JOIN public.user_profile_category upc ON up.id = upc.user_profile_id
  WHERE up.gemini_score IS NOT NULL AND upc.category_id = $1
  GROUP BY score_range
),
model_counts AS (
  SELECT
    CASE 
      WHEN up.model_score BETWEEN 0.0 AND 0.2 THEN '0.0-0.2'
      WHEN up.model_score BETWEEN 0.2 AND 0.4 THEN '0.2-0.4'
      WHEN up.model_score BETWEEN 0.4 AND 0.6 THEN '0.4-0.6'
      WHEN up.model_score BETWEEN 0.6 AND 0.8 THEN '0.6-0.8'
      WHEN up.model_score BETWEEN 0.8 AND 1.0 THEN '0.8-1.0'
    END as score_range,
    COUNT(*) as model_count
  FROM public.user_profile up
  JOIN public.user_profile_category upc ON up.id = upc.user_profile_id
  WHERE up.model_score IS NOT NULL AND upc.category_id = $1
  GROUP BY score_range
)
SELECT 
  sr.range as score_range,
  COALESCE(gc.gemini_count, 0) as gemini_count,
  COALESCE(mc.model_count, 0) as model_count
FROM score_ranges sr
LEFT JOIN gemini_counts gc ON sr.range = gc.score_range
LEFT JOIN model_counts mc ON sr.range = mc.score_range
ORDER BY sr.range;

-- name: GetCategories :many
SELECT * FROM public.category ORDER BY name;

-- name: CreateCategory :one
INSERT INTO public.category (name, description, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM public.category WHERE id = $1;

-- name: UpdateCategory :one
UPDATE public.category
SET name = $2,
    description = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM public.category WHERE id = $1;

-- name: CountProfilesInCategory :one
SELECT COUNT(*) FROM public.user_profile_category WHERE category_id = $1;

-- name: AddAllProfilesToCategory :execrows
INSERT INTO public.user_profile_category (user_profile_id, category_id, created_at)
SELECT up.id, $1, NOW()
FROM public.user_profile up
WHERE up.is_scanned = true
AND NOT EXISTS (
    SELECT 1 FROM public.user_profile_category upc
    WHERE upc.user_profile_id = up.id AND upc.category_id = $1
);
-- Model Management queries
-- name: GetModels :many
SELECT * FROM public.model
ORDER BY created_at DESC;

-- name: CreateModel :one
INSERT INTO public.model (name, description, category_id, created_at)
VALUES ($1, $2, $3, NOW())
RETURNING *;

-- name: UpdateModel :one
UPDATE public.model
SET name = $2,
    description = $3,
    category_id = $4
WHERE id = $1
RETURNING *;

-- name: DeleteModel :exec
DELETE FROM public.model WHERE id = $1;

-- name: GetModelByID :one
SELECT * FROM public.model WHERE id = $1;

-- name: GetModelByCategory :one
SELECT * FROM public.model WHERE category_id = $1;

-- name: GetModelsWithoutCategory :many
SELECT * FROM public.model WHERE category_id IS NULL ORDER BY created_at DESC;