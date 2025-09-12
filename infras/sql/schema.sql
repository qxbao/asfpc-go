CREATE TABLE IF NOT EXISTS public.proxy
(
    id integer NOT NULL DEFAULT nextval('proxy_id_seq'::regclass),
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
    id integer NOT NULL DEFAULT nextval('account_id_seq'::regclass),
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
    CONSTRAINT account_email_key UNIQUE (email),
    CONSTRAINT account_username_key UNIQUE (username),
    CONSTRAINT account_proxy_id_fkey FOREIGN KEY (proxy_id)
        REFERENCES public.proxy (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS public."group"
(
    id integer NOT NULL DEFAULT nextval('group_id_seq'::regclass),
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
    id integer NOT NULL DEFAULT nextval('post_id_seq'::regclass),
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

CREATE TABLE IF NOT EXISTS public.comment
(
    id integer NOT NULL DEFAULT nextval('comment_id_seq'::regclass),
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

CREATE TABLE IF NOT EXISTS public.user_profile
(
    id integer NOT NULL DEFAULT nextval('user_profile_id_seq'::regclass),
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
    profile_url character varying COLLATE pg_catalog."default" NOT NULL DEFAULT 'NOT SPECIFIED'::character varying,
    CONSTRAINT user_profile_pkey PRIMARY KEY (id),
    CONSTRAINT user_profile_facebook_id_key UNIQUE (facebook_id),
    CONSTRAINT user_profile_scraped_by_id_fkey FOREIGN KEY (scraped_by_id)
        REFERENCES public.account (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS public.image
(
    id integer NOT NULL DEFAULT nextval('image_id_seq'::regclass),
    path character varying COLLATE pg_catalog."default" NOT NULL,
    is_analyzed boolean NOT NULL,
    belong_to_id integer NOT NULL,
    CONSTRAINT image_pkey PRIMARY KEY (id),
    CONSTRAINT image_belong_to_id_fkey FOREIGN KEY (belong_to_id)
        REFERENCES public.user_profile (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS public.prompt
(
    id integer NOT NULL DEFAULT nextval('prompt_id_seq'::regclass),
    content character varying COLLATE pg_catalog."default" NOT NULL,
    service_name character varying COLLATE pg_catalog."default" NOT NULL,
    version integer NOT NULL,
    created_by character varying COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp without time zone NOT NULL,
    CONSTRAINT prompt_pkey PRIMARY KEY (id),
    CONSTRAINT uq_service_name_version UNIQUE (service_name, version)
);

CREATE TABLE IF NOT EXISTS public.financial_analysis
(
    id integer NOT NULL DEFAULT nextval('financial_analysis_id_seq'::regclass),
    financial_status character varying COLLATE pg_catalog."default" NOT NULL,
    confidence_score double precision NOT NULL,
    analysis_summary text COLLATE pg_catalog."default" NOT NULL,
    indicators json,
    gemini_model_used character varying COLLATE pg_catalog."default" NOT NULL,
    prompt_tokens_used integer,
    prompt_used_id integer NOT NULL,
    completion_tokens_used integer,
    total_tokens_used integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    user_profile_id integer NOT NULL,
    CONSTRAINT financial_analysis_pkey PRIMARY KEY (id),
    CONSTRAINT financial_analysis_prompt_used_id_fkey FOREIGN KEY (prompt_used_id)
        REFERENCES public.prompt (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT financial_analysis_user_profile_id_fkey FOREIGN KEY (user_profile_id)
        REFERENCES public.user_profile (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS public.config
(
    id integer NOT NULL DEFAULT nextval('config_id_seq'::regclass),
    key character varying COLLATE pg_catalog."default" NOT NULL,
    value character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT config_pkey PRIMARY KEY (id)
);
