-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.model
(
    id SERIAL,
    name character varying COLLATE pg_catalog."default" NOT NULL,
    description character varying COLLATE pg_catalog."default",
    created_at timestamp without time zone NOT NULL,
    category_id integer,
    CONSTRAINT model_pkey PRIMARY KEY (id),
    CONSTRAINT uq_model_name UNIQUE (name),
    CONSTRAINT model_category_id_fkey FOREIGN KEY (category_id)
        REFERENCES public.category (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE SET NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_model_category_id ON public.model (category_id) WHERE category_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS public.uq_model_category_id;
DROP TABLE IF EXISTS public.model;
-- +goose StatementEnd
