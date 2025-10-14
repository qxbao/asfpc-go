--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5
-- Dumped by pg_dump version 17.5

-- Started on 2025-10-14 11:00:03

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 6 (class 2615 OID 33423)
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

-- *not* creating schema, since initdb creates it


--
-- TOC entry 2 (class 3079 OID 33816)
-- Name: vector; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 225 (class 1259 OID 33448)
-- Name: account; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.account (
    id integer NOT NULL,
    email character varying NOT NULL,
    username character varying NOT NULL,
    password character varying NOT NULL,
    is_block boolean NOT NULL,
    ua character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    cookies json,
    access_token character varying,
    proxy_id integer
);


--
-- TOC entry 224 (class 1259 OID 33447)
-- Name: account_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.account_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5309 (class 0 OID 0)
-- Dependencies: 224
-- Name: account_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.account_id_seq OWNED BY public.account.id;


--
-- TOC entry 249 (class 1259 OID 89601)
-- Name: category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.category (
    id integer NOT NULL,
    name character varying NOT NULL,
    description character varying,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- TOC entry 248 (class 1259 OID 89600)
-- Name: category_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.category_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5310 (class 0 OID 0)
-- Dependencies: 248
-- Name: category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.category_id_seq OWNED BY public.category.id;


--
-- TOC entry 231 (class 1259 OID 33496)
-- Name: comment; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.comment (
    id integer NOT NULL,
    content character varying NOT NULL,
    is_analyzed boolean NOT NULL,
    created_at timestamp without time zone NOT NULL,
    inserted_at timestamp without time zone NOT NULL,
    post_id integer NOT NULL,
    author_id integer NOT NULL,
    comment_id character varying NOT NULL
);


--
-- TOC entry 230 (class 1259 OID 33495)
-- Name: comment_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.comment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5311 (class 0 OID 0)
-- Dependencies: 230
-- Name: comment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.comment_id_seq OWNED BY public.comment.id;


--
-- TOC entry 237 (class 1259 OID 33605)
-- Name: config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.config (
    id integer NOT NULL,
    key character varying NOT NULL,
    value character varying NOT NULL
);


--
-- TOC entry 236 (class 1259 OID 33604)
-- Name: config_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5312 (class 0 OID 0)
-- Dependencies: 236
-- Name: config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.config_id_seq OWNED BY public.config.id;


--
-- TOC entry 243 (class 1259 OID 34191)
-- Name: embedded_profile; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.embedded_profile (
    id integer NOT NULL,
    pid integer NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    embedding public.vector(1024)
);


--
-- TOC entry 242 (class 1259 OID 34190)
-- Name: embedded_profile_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.embedded_profile_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5313 (class 0 OID 0)
-- Dependencies: 242
-- Name: embedded_profile_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.embedded_profile_id_seq OWNED BY public.embedded_profile.id;


--
-- TOC entry 241 (class 1259 OID 33775)
-- Name: gemini_key; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gemini_key (
    id integer NOT NULL,
    api_key text NOT NULL,
    token_used bigint DEFAULT 0 NOT NULL,
    updated_at timestamp without time zone DEFAULT now()
);


--
-- TOC entry 240 (class 1259 OID 33774)
-- Name: gemini_key_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.gemini_key_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5314 (class 0 OID 0)
-- Dependencies: 240
-- Name: gemini_key_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.gemini_key_id_seq OWNED BY public.gemini_key.id;


--
-- TOC entry 247 (class 1259 OID 89594)
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now() NOT NULL
);


--
-- TOC entry 246 (class 1259 OID 89593)
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.goose_db_version ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.goose_db_version_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 227 (class 1259 OID 33466)
-- Name: group; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public."group" (
    id integer NOT NULL,
    group_id character varying NOT NULL,
    group_name character varying NOT NULL,
    is_joined boolean NOT NULL,
    account_id integer,
    scanned_at timestamp without time zone
);


--
-- TOC entry 250 (class 1259 OID 89611)
-- Name: group_category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_category (
    group_id integer NOT NULL,
    category_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- TOC entry 226 (class 1259 OID 33465)
-- Name: group_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5315 (class 0 OID 0)
-- Dependencies: 226
-- Name: group_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_id_seq OWNED BY public."group".id;


--
-- TOC entry 239 (class 1259 OID 33627)
-- Name: log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.log (
    id integer NOT NULL,
    account_id integer,
    action character varying NOT NULL,
    target_id integer,
    description text,
    created_at timestamp without time zone DEFAULT now()
);


--
-- TOC entry 238 (class 1259 OID 33626)
-- Name: log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5316 (class 0 OID 0)
-- Dependencies: 238
-- Name: log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.log_id_seq OWNED BY public.log.id;


--
-- TOC entry 229 (class 1259 OID 33482)
-- Name: post; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.post (
    id integer NOT NULL,
    post_id character varying NOT NULL,
    content character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    inserted_at timestamp without time zone NOT NULL,
    group_id integer NOT NULL,
    is_analyzed boolean NOT NULL,
    scanned_at timestamp without time zone
);


--
-- TOC entry 228 (class 1259 OID 33481)
-- Name: post_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5317 (class 0 OID 0)
-- Dependencies: 228
-- Name: post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.post_id_seq OWNED BY public.post.id;


--
-- TOC entry 233 (class 1259 OID 33535)
-- Name: prompt; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.prompt (
    id integer NOT NULL,
    content character varying NOT NULL,
    service_name character varying NOT NULL,
    version integer NOT NULL,
    created_by character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    category_id integer NOT NULL
);


--
-- TOC entry 232 (class 1259 OID 33534)
-- Name: prompt_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.prompt_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5318 (class 0 OID 0)
-- Dependencies: 232
-- Name: prompt_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.prompt_id_seq OWNED BY public.prompt.id;


--
-- TOC entry 223 (class 1259 OID 33439)
-- Name: proxy; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.proxy (
    id integer NOT NULL,
    ip character varying NOT NULL,
    port character varying NOT NULL,
    username character varying NOT NULL,
    password character varying NOT NULL,
    is_active boolean NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- TOC entry 222 (class 1259 OID 33438)
-- Name: proxy_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.proxy_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5319 (class 0 OID 0)
-- Dependencies: 222
-- Name: proxy_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.proxy_id_seq OWNED BY public.proxy.id;


--
-- TOC entry 245 (class 1259 OID 65238)
-- Name: request; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.request (
    id integer NOT NULL,
    progress double precision DEFAULT 0 NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    description character varying(50),
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    error_message text
);


--
-- TOC entry 244 (class 1259 OID 65237)
-- Name: request_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.request_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5320 (class 0 OID 0)
-- Dependencies: 244
-- Name: request_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.request_id_seq OWNED BY public.request.id;


--
-- TOC entry 235 (class 1259 OID 33546)
-- Name: user_profile; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_profile (
    id integer NOT NULL,
    facebook_id character varying NOT NULL,
    name character varying,
    bio text,
    location character varying,
    work text,
    education text,
    relationship_status character varying,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    scraped_by_id integer NOT NULL,
    is_scanned boolean DEFAULT false NOT NULL,
    hometown character varying,
    locale character varying(16) DEFAULT 'NOT_SPECIFIED'::character varying NOT NULL,
    gender character varying(16),
    birthday character varying(10),
    email character varying(100),
    phone character varying(12),
    profile_url character varying DEFAULT 'NOT_SPECIFIED'::character varying NOT NULL,
    is_analyzed boolean DEFAULT false,
    gemini_score double precision,
    model_score double precision
);


--
-- TOC entry 251 (class 1259 OID 89627)
-- Name: user_profile_category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_profile_category (
    user_profile_id integer NOT NULL,
    category_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- TOC entry 234 (class 1259 OID 33545)
-- Name: user_profile_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_profile_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5321 (class 0 OID 0)
-- Dependencies: 234
-- Name: user_profile_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_profile_id_seq OWNED BY public.user_profile.id;


--
-- TOC entry 5058 (class 2604 OID 33451)
-- Name: account id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account ALTER COLUMN id SET DEFAULT nextval('public.account_id_seq'::regclass);


--
-- TOC entry 5082 (class 2604 OID 89604)
-- Name: category id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.category ALTER COLUMN id SET DEFAULT nextval('public.category_id_seq'::regclass);


--
-- TOC entry 5061 (class 2604 OID 33499)
-- Name: comment id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comment ALTER COLUMN id SET DEFAULT nextval('public.comment_id_seq'::regclass);


--
-- TOC entry 5068 (class 2604 OID 33608)
-- Name: config id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.config ALTER COLUMN id SET DEFAULT nextval('public.config_id_seq'::regclass);


--
-- TOC entry 5074 (class 2604 OID 34194)
-- Name: embedded_profile id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.embedded_profile ALTER COLUMN id SET DEFAULT nextval('public.embedded_profile_id_seq'::regclass);


--
-- TOC entry 5071 (class 2604 OID 33778)
-- Name: gemini_key id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gemini_key ALTER COLUMN id SET DEFAULT nextval('public.gemini_key_id_seq'::regclass);


--
-- TOC entry 5059 (class 2604 OID 33469)
-- Name: group id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."group" ALTER COLUMN id SET DEFAULT nextval('public.group_id_seq'::regclass);


--
-- TOC entry 5069 (class 2604 OID 33630)
-- Name: log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log ALTER COLUMN id SET DEFAULT nextval('public.log_id_seq'::regclass);


--
-- TOC entry 5060 (class 2604 OID 33485)
-- Name: post id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.post ALTER COLUMN id SET DEFAULT nextval('public.post_id_seq'::regclass);


--
-- TOC entry 5062 (class 2604 OID 33538)
-- Name: prompt id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prompt ALTER COLUMN id SET DEFAULT nextval('public.prompt_id_seq'::regclass);


--
-- TOC entry 5057 (class 2604 OID 33442)
-- Name: proxy id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.proxy ALTER COLUMN id SET DEFAULT nextval('public.proxy_id_seq'::regclass);


--
-- TOC entry 5076 (class 2604 OID 65241)
-- Name: request id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.request ALTER COLUMN id SET DEFAULT nextval('public.request_id_seq'::regclass);


--
-- TOC entry 5063 (class 2604 OID 33549)
-- Name: user_profile id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profile ALTER COLUMN id SET DEFAULT nextval('public.user_profile_id_seq'::regclass);


--
-- TOC entry 5090 (class 2606 OID 33455)
-- Name: account account_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT account_pkey PRIMARY KEY (id);


--
-- TOC entry 5092 (class 2606 OID 33459)
-- Name: account account_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT account_username_key UNIQUE (username);


--
-- TOC entry 5137 (class 2606 OID 89608)
-- Name: category category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_pkey PRIMARY KEY (id);


--
-- TOC entry 5102 (class 2606 OID 33621)
-- Name: comment comment_comment_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comment
    ADD CONSTRAINT comment_comment_id_key UNIQUE (comment_id);


--
-- TOC entry 5104 (class 2606 OID 33503)
-- Name: comment comment_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comment
    ADD CONSTRAINT comment_pkey PRIMARY KEY (id);


--
-- TOC entry 5118 (class 2606 OID 64769)
-- Name: config config_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.config
    ADD CONSTRAINT config_key_key UNIQUE (key);


--
-- TOC entry 5120 (class 2606 OID 33612)
-- Name: config config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.config
    ADD CONSTRAINT config_pkey PRIMARY KEY (id);


--
-- TOC entry 5129 (class 2606 OID 34201)
-- Name: embedded_profile embedded_profile_pid_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.embedded_profile
    ADD CONSTRAINT embedded_profile_pid_key UNIQUE (pid);


--
-- TOC entry 5131 (class 2606 OID 34199)
-- Name: embedded_profile embedded_profile_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.embedded_profile
    ADD CONSTRAINT embedded_profile_pkey PRIMARY KEY (id);


--
-- TOC entry 5126 (class 2606 OID 33783)
-- Name: gemini_key gemini_key_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gemini_key
    ADD CONSTRAINT gemini_key_pkey PRIMARY KEY (id);


--
-- TOC entry 5135 (class 2606 OID 89599)
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- TOC entry 5141 (class 2606 OID 89616)
-- Name: group_category group_category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_category
    ADD CONSTRAINT group_category_pkey PRIMARY KEY (group_id, category_id);


--
-- TOC entry 5094 (class 2606 OID 33473)
-- Name: group group_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."group"
    ADD CONSTRAINT group_pkey PRIMARY KEY (id);


--
-- TOC entry 5124 (class 2606 OID 33635)
-- Name: log log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log
    ADD CONSTRAINT log_pkey PRIMARY KEY (id);


--
-- TOC entry 5098 (class 2606 OID 33489)
-- Name: post post_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT post_pkey PRIMARY KEY (id);


--
-- TOC entry 5100 (class 2606 OID 33603)
-- Name: post post_post_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT post_post_id_key UNIQUE (post_id);


--
-- TOC entry 5106 (class 2606 OID 33542)
-- Name: prompt prompt_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prompt
    ADD CONSTRAINT prompt_pkey PRIMARY KEY (id);


--
-- TOC entry 5086 (class 2606 OID 33446)
-- Name: proxy proxy_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.proxy
    ADD CONSTRAINT proxy_pkey PRIMARY KEY (id);


--
-- TOC entry 5133 (class 2606 OID 65245)
-- Name: request request_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.request
    ADD CONSTRAINT request_pkey PRIMARY KEY (id);


--
-- TOC entry 5139 (class 2606 OID 89610)
-- Name: category uq_category_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT uq_category_name UNIQUE (name);


--
-- TOC entry 5096 (class 2606 OID 33528)
-- Name: group uq_group_account; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."group"
    ADD CONSTRAINT uq_group_account UNIQUE (group_id, account_id);


--
-- TOC entry 5088 (class 2606 OID 33511)
-- Name: proxy uq_ip_port_username; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.proxy
    ADD CONSTRAINT uq_ip_port_username UNIQUE (ip, port, username);


--
-- TOC entry 5108 (class 2606 OID 89644)
-- Name: prompt uq_prompt_service_name_category; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prompt
    ADD CONSTRAINT uq_prompt_service_name_category UNIQUE (service_name, category_id);


--
-- TOC entry 5110 (class 2606 OID 89665)
-- Name: prompt uq_service_category_version; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prompt
    ADD CONSTRAINT uq_service_category_version UNIQUE (service_name, category_id, version);


--
-- TOC entry 5112 (class 2606 OID 33544)
-- Name: prompt uq_service_name_version; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prompt
    ADD CONSTRAINT uq_service_name_version UNIQUE (service_name, version);


--
-- TOC entry 5145 (class 2606 OID 89632)
-- Name: user_profile_category user_profile_category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profile_category
    ADD CONSTRAINT user_profile_category_pkey PRIMARY KEY (user_profile_id, category_id);


--
-- TOC entry 5114 (class 2606 OID 33555)
-- Name: user_profile user_profile_facebook_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profile
    ADD CONSTRAINT user_profile_facebook_id_key UNIQUE (facebook_id);


--
-- TOC entry 5116 (class 2606 OID 33553)
-- Name: user_profile user_profile_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profile
    ADD CONSTRAINT user_profile_pkey PRIMARY KEY (id);


--
-- TOC entry 5127 (class 1259 OID 71054)
-- Name: embedded_profile_embedding_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX embedded_profile_embedding_idx ON public.embedded_profile USING ivfflat (embedding public.vector_cosine_ops) WITH (lists='100');


--
-- TOC entry 5142 (class 1259 OID 89650)
-- Name: idx_group_category_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_group_category_category_id ON public.group_category USING btree (category_id);


--
-- TOC entry 5143 (class 1259 OID 89651)
-- Name: idx_user_profile_category_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_profile_category_category_id ON public.user_profile_category USING btree (category_id);


--
-- TOC entry 5121 (class 1259 OID 33613)
-- Name: ix_config_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_config_id ON public.config USING btree (id);


--
-- TOC entry 5122 (class 1259 OID 33614)
-- Name: ix_config_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ix_config_key ON public.config USING btree (key);


--
-- TOC entry 5146 (class 2606 OID 33460)
-- Name: account account_proxy_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT account_proxy_id_fkey FOREIGN KEY (proxy_id) REFERENCES public.proxy(id);


--
-- TOC entry 5149 (class 2606 OID 33615)
-- Name: comment comment_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comment
    ADD CONSTRAINT comment_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.user_profile(id);


--
-- TOC entry 5150 (class 2606 OID 33504)
-- Name: comment comment_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comment
    ADD CONSTRAINT comment_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.post(id);


--
-- TOC entry 5154 (class 2606 OID 34202)
-- Name: embedded_profile embedded_profile_pid_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.embedded_profile
    ADD CONSTRAINT embedded_profile_pid_fk FOREIGN KEY (pid) REFERENCES public.user_profile(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5147 (class 2606 OID 33529)
-- Name: group group_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."group"
    ADD CONSTRAINT group_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(id);


--
-- TOC entry 5155 (class 2606 OID 89622)
-- Name: group_category group_category_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_category
    ADD CONSTRAINT group_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.category(id) ON DELETE CASCADE;


--
-- TOC entry 5156 (class 2606 OID 89617)
-- Name: group_category group_category_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_category
    ADD CONSTRAINT group_category_group_id_fkey FOREIGN KEY (group_id) REFERENCES public."group"(id) ON DELETE CASCADE;


--
-- TOC entry 5153 (class 2606 OID 33636)
-- Name: log log_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log
    ADD CONSTRAINT log_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5148 (class 2606 OID 33490)
-- Name: post post_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT post_group_id_fkey FOREIGN KEY (group_id) REFERENCES public."group"(id);


--
-- TOC entry 5151 (class 2606 OID 89645)
-- Name: prompt prompt_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prompt
    ADD CONSTRAINT prompt_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.category(id) ON DELETE CASCADE;


--
-- TOC entry 5157 (class 2606 OID 89638)
-- Name: user_profile_category user_profile_category_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profile_category
    ADD CONSTRAINT user_profile_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.category(id) ON DELETE CASCADE;


--
-- TOC entry 5158 (class 2606 OID 89633)
-- Name: user_profile_category user_profile_category_user_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profile_category
    ADD CONSTRAINT user_profile_category_user_profile_id_fkey FOREIGN KEY (user_profile_id) REFERENCES public.user_profile(id) ON DELETE CASCADE;


--
-- TOC entry 5152 (class 2606 OID 33556)
-- Name: user_profile user_profile_scraped_by_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profile
    ADD CONSTRAINT user_profile_scraped_by_id_fkey FOREIGN KEY (scraped_by_id) REFERENCES public.account(id);


-- Completed on 2025-10-14 11:00:03

--
-- PostgreSQL database dump complete
--

