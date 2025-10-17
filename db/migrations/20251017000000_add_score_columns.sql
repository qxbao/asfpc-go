-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.user_profile 
ADD COLUMN IF NOT EXISTS gemini_score double precision,
ADD COLUMN IF NOT EXISTS model_score double precision;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.user_profile 
DROP COLUMN IF EXISTS gemini_score,
DROP COLUMN IF EXISTS model_score;
-- +goose StatementEnd
