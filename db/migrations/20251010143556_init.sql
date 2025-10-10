-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.proxy
(
    id SERIAL,
    ip character varying COLLATE pg_catalog."default" NOT NULL,
    port character varying COLLATE pg_catalog."default" NOT NULL,
    username character varying COLLATE pg_catalog."default" NOT NULL,
    password character varying COLLATE pg_catalog."default" NOT NULL,
    is_active boolean NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    CONSTRAINT proxy_pkey PRIMARY KEY (id),
    CONSTRAINT uq_ip_port_username UNIQUE (ip, port, username)
);

CREATE TABLE IF NOT EXISTS public.account
(
    id SERIAL,
    email character varying COLLATE pg_catalog."default" NOT NULL,
    username character varying COLLATE pg_catalog."default" NOT NULL,
    password character varying COLLATE pg_catalog."default" NOT NULL,
    is_block boolean NOT NULL,
    ua character varying COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    cookies json,
    access_token character varying COLLATE pg_catalog."default",
    proxy_id integer,
    CONSTRAINT account_pkey PRIMARY KEY (id),
    CONSTRAINT account_username_key UNIQUE (username),
    CONSTRAINT account_proxy_id_fkey FOREIGN KEY (proxy_id)
        REFERENCES public.proxy (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS public."group"
(
    id SERIAL,
    group_id character varying COLLATE pg_catalog."default" NOT NULL,
    group_name character varying COLLATE pg_catalog."default" NOT NULL,
    is_joined boolean NOT NULL,
    account_id integer,
    scanned_at timestamp without time zone,
    CONSTRAINT group_pkey PRIMARY KEY (id),
    CONSTRAINT uq_group_account UNIQUE (group_id, account_id),
    CONSTRAINT group_account_id_fkey FOREIGN KEY (account_id)
        REFERENCES public.account (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS public.post
(
    id SERIAL,
    post_id character varying COLLATE pg_catalog."default" NOT NULL,
    content character varying COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp without time zone NOT NULL,
    inserted_at timestamp without time zone NOT NULL,
    group_id integer NOT NULL,
    is_analyzed boolean NOT NULL,
    CONSTRAINT post_pkey PRIMARY KEY (id),
    CONSTRAINT post_post_id_key UNIQUE (post_id),
    CONSTRAINT post_group_id_fkey FOREIGN KEY (group_id)
        REFERENCES public."group" (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS public.user_profile
(
    id SERIAL,
    facebook_id character varying COLLATE pg_catalog."default" NOT NULL,
    name character varying COLLATE pg_catalog."default",
    bio text COLLATE pg_catalog."default",
    location character varying COLLATE pg_catalog."default",
    work text COLLATE pg_catalog."default",
    education text COLLATE pg_catalog."default",
    relationship_status character varying COLLATE pg_catalog."default",
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    scraped_by_id integer NOT NULL,
    is_scanned boolean NOT NULL DEFAULT false,
    hometown character varying COLLATE pg_catalog."default",
    locale character varying(16) COLLATE pg_catalog."default" NOT NULL DEFAULT 'NOT_SPECIFIED'::character varying,
    gender character varying(16) COLLATE pg_catalog."default",
    birthday character varying(10) COLLATE pg_catalog."default",
    email character varying(100) COLLATE pg_catalog."default",
    phone character varying(12) COLLATE pg_catalog."default",
    profile_url character varying COLLATE pg_catalog."default" NOT NULL DEFAULT 'NOT_SPECIFIED'::character varying,
    is_analyzed boolean DEFAULT false,
    gemini_score double precision,
    model_score double precision,
    CONSTRAINT user_profile_pkey PRIMARY KEY (id),
    CONSTRAINT user_profile_facebook_id_key UNIQUE (facebook_id),
    CONSTRAINT user_profile_scraped_by_id_fkey FOREIGN KEY (scraped_by_id)
        REFERENCES public.account (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS public.comment
(
    id SERIAL,
    content character varying COLLATE pg_catalog."default" NOT NULL,
    is_analyzed boolean NOT NULL,
    created_at timestamp without time zone NOT NULL,
    inserted_at timestamp without time zone NOT NULL,
    post_id integer NOT NULL,
    author_id integer NOT NULL,
    comment_id character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT comment_pkey PRIMARY KEY (id),
    CONSTRAINT comment_comment_id_key UNIQUE (comment_id),
    CONSTRAINT comment_author_id_fkey FOREIGN KEY (author_id)
        REFERENCES public.user_profile (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT comment_post_id_fkey FOREIGN KEY (post_id)
        REFERENCES public.post (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS public.prompt
(
    id SERIAL,
    content character varying COLLATE pg_catalog."default" NOT NULL,
    service_name character varying COLLATE pg_catalog."default" NOT NULL,
    version integer NOT NULL,
    created_by character varying COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp without time zone NOT NULL,
    CONSTRAINT prompt_pkey PRIMARY KEY (id),
    CONSTRAINT uq_service_name_version UNIQUE (service_name, version)
);

CREATE TABLE IF NOT EXISTS public.config
(
    id SERIAL,
    key character varying COLLATE pg_catalog."default" NOT NULL,
    value character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT config_key_key UNIQUE (key),
    CONSTRAINT config_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.log
(
    id SERIAL,
    account_id integer,
    action character varying COLLATE pg_catalog."default" NOT NULL,
    target_id integer,
    description text COLLATE pg_catalog."default",
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT log_pkey PRIMARY KEY (id),
    CONSTRAINT log_account_id_fkey FOREIGN KEY (account_id)
        REFERENCES public.account (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.gemini_key
(
    id SERIAL,
    api_key text COLLATE pg_catalog."default" NOT NULL,
    token_used bigint NOT NULL DEFAULT 0,
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT gemini_key_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.embedded_profile
(
    id SERIAL,
    pid integer NOT NULL,
    embedding vector(1024),
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT embedded_profile_pkey PRIMARY KEY (id),
    CONSTRAINT embedded_profile_pid_key UNIQUE(pid),
    CONSTRAINT embedded_profile_pid_fk FOREIGN KEY (pid)
        REFERENCES public.user_profile (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS public.request
(
    id SERIAL,
    progress double precision NOT NULL DEFAULT 0,
    status smallint NOT NULL DEFAULT 0, 
    description character varying(50) COLLATE pg_catalog."default", 
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    error_message text COLLATE pg_catalog."default",
    CONSTRAINT request_pkey PRIMARY KEY (id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.request;
DROP TABLE IF EXISTS public.embedded_profile;
DROP TABLE IF EXISTS public.gemini_key;
DROP TABLE IF EXISTS public.log;
DROP TABLE IF EXISTS public.config;
DROP TABLE IF EXISTS public.prompt;
DROP TABLE IF EXISTS public.comment;
DROP TABLE IF EXISTS public.user_profile;
DROP TABLE IF EXISTS public.post;
DROP TABLE IF EXISTS public."group";
DROP TABLE IF EXISTS public.account;
DROP TABLE IF EXISTS public.proxy;
-- +goose StatementEnd
