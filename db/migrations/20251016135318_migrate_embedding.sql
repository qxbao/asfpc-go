-- +goose Up
-- +goose StatementBegin

ALTER TABLE public.embedded_profile
  ADD COLUMN IF NOT EXISTS cid integer,
  DROP CONSTRAINT IF EXISTS embedded_profile_pid_key,
  DROP CONSTRAINT IF EXISTS embedded_profile_pid_fk;

UPDATE public.embedded_profile ep
SET cid = (
  SELECT upc.category_id 
  FROM public.user_profile_category upc 
  WHERE upc.user_profile_id = ep.pid 
  LIMIT 1
)
WHERE ep.cid IS NULL;

ALTER TABLE public.embedded_profile
  ALTER COLUMN cid SET NOT NULL,
  ADD CONSTRAINT embedded_profile_pid_key UNIQUE(pid, cid),
  ADD CONSTRAINT embedded_profile_category_id_fkey FOREIGN KEY (cid)
    REFERENCES public.category (id) MATCH SIMPLE
    ON UPDATE CASCADE
    ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.embedded_profile
  DROP CONSTRAINT IF EXISTS embedded_profile_pid_key,
  DROP CONSTRAINT IF EXISTS embedded_profile_category_id_fkey,
  DROP COLUMN IF EXISTS cid;
ALTER TABLE public.embedded_profile
  ADD CONSTRAINT embedded_profile_pid_key UNIQUE(pid);
ALTER TABLE public.embedded_profile
  ADD CONSTRAINT embedded_profile_pid_fk FOREIGN KEY (pid)
    REFERENCES public.user_profile (id) MATCH SIMPLE
    ON UPDATE CASCADE
    ON DELETE CASCADE;
-- +goose StatementEnd
