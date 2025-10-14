-- +goose Up
-- +goose StatementBegin
DELETE FROM public.prompt WHERE category_id IS NULL;
ALTER TABLE public.prompt
ADD CONSTRAINT uq_service_category_version UNIQUE (service_name, category_id, version),
ALTER COLUMN category_id SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.prompt
DROP CONSTRAINT uq_service_category_version,
ALTER COLUMN category_id DROP NOT NULL;
-- +goose StatementEnd
