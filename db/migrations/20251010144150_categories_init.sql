-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.category
(
    id SERIAL,
    name character varying COLLATE pg_catalog."default" NOT NULL,
    description character varying COLLATE pg_catalog."default",
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    CONSTRAINT category_pkey PRIMARY KEY (id),
    CONSTRAINT uq_category_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS public.group_category
(
    group_id integer NOT NULL,
    category_id integer NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    CONSTRAINT group_category_pkey PRIMARY KEY (group_id, category_id),
    CONSTRAINT group_category_group_id_fkey FOREIGN KEY (group_id)
        REFERENCES public."group" (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE,
    CONSTRAINT group_category_category_id_fkey FOREIGN KEY (category_id)
        REFERENCES public.category (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.user_profile_category
(
    user_profile_id integer NOT NULL,
    category_id integer NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    CONSTRAINT user_profile_category_pkey PRIMARY KEY (user_profile_id, category_id),
    CONSTRAINT user_profile_category_user_profile_id_fkey FOREIGN KEY (user_profile_id)
        REFERENCES public.user_profile (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE,
    CONSTRAINT user_profile_category_category_id_fkey FOREIGN KEY (category_id)
        REFERENCES public.category (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

ALTER TABLE public."prompt"
  ADD COLUMN IF NOT EXISTS category_id integer,
  ADD CONSTRAINT uq_prompt_service_name_category UNIQUE (service_name, category_id),
  ADD CONSTRAINT prompt_category_id_fkey FOREIGN KEY (category_id)
        REFERENCES public.category (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_group_category_category_id ON public.group_category(category_id);
CREATE INDEX IF NOT EXISTS idx_user_profile_category_category_id ON public.user_profile_category(category_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS public.idx_user_profile_category_category_id;
DROP INDEX IF EXISTS public.idx_group_category_category_id;

DROP TABLE IF EXISTS public.user_profile_category;
DROP TABLE IF EXISTS public.group_category;

ALTER TABLE public."prompt"
  DROP CONSTRAINT IF EXISTS uq_prompt_service_name_category,
  DROP CONSTRAINT IF EXISTS prompt_category_id_fkey,
  DROP COLUMN IF EXISTS category_id;

DROP TABLE IF EXISTS public.category;
-- +goose StatementEnd