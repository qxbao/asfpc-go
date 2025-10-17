-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.user_profile_category
  ADD COLUMN IF NOT EXISTS model_score double precision,
  ADD COLUMN IF NOT EXISTS gemini_score double precision;

UPDATE public.user_profile_category upc
SET 
  model_score = up.model_score,
  gemini_score = up.gemini_score
FROM public.user_profile up
WHERE upc.user_profile_id = up.id
  AND (up.model_score IS NOT NULL OR up.gemini_score IS NOT NULL);

ALTER TABLE public.user_profile
  DROP COLUMN IF EXISTS model_score,
  DROP COLUMN IF EXISTS gemini_score;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.user_profile
  ADD COLUMN IF NOT EXISTS model_score double precision,
  ADD COLUMN IF NOT EXISTS gemini_score double precision;

UPDATE public.user_profile up
SET 
  model_score = (
    SELECT upc.model_score 
    FROM public.user_profile_category upc 
    WHERE upc.user_profile_id = up.id 
    LIMIT 1
  ),
  gemini_score = (
    SELECT upc.gemini_score 
    FROM public.user_profile_category upc 
    WHERE upc.user_profile_id = up.id 
    LIMIT 1
  );

ALTER TABLE public.user_profile_category
  DROP COLUMN IF EXISTS model_score,
  DROP COLUMN IF EXISTS gemini_score;
-- +goose StatementEnd
