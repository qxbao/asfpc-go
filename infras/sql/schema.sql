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
)

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
)