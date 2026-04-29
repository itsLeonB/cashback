--
-- PostgreSQL database dump
--

\restrict l53tlQ1rF2r6rqVgGl707x5be7yFSnrHUzLlRQ72v1pbmgWEhUdMSdg0TF1irJD

-- Dumped from database version 18.1 (Ubuntu 18.1-1.pgdg24.04+2)
-- Dumped by pg_dump version 18.1 (Ubuntu 18.1-1.pgdg24.04+2)

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
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: postgres_fdw; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgres_fdw WITH SCHEMA public;


--
-- Name: EXTENSION postgres_fdw; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION postgres_fdw IS 'foreign-data wrapper for remote PostgreSQL servers';


--
-- Name: fee_calculation_method; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.fee_calculation_method AS ENUM (
    'EQUAL_SPLIT',
    'ITEMIZED_SPLIT'
);


ALTER TYPE public.fee_calculation_method OWNER TO postgres;

--
-- Name: friendship_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.friendship_type AS ENUM (
    'REAL',
    'ANON'
);


ALTER TYPE public.friendship_type OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin_roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.admin_roles (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name text NOT NULL
);


ALTER TABLE public.admin_roles OWNER TO postgres;

--
-- Name: admin_users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.admin_users (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    email text NOT NULL,
    password text NOT NULL
);


ALTER TABLE public.admin_users OWNER TO postgres;

--
-- Name: admin_users_roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.admin_users_roles (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    user_id uuid NOT NULL,
    role_id uuid NOT NULL
);


ALTER TABLE public.admin_users_roles OWNER TO postgres;

--
-- Name: debt_transactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.debt_transactions (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    lender_profile_id uuid NOT NULL,
    borrower_profile_id uuid NOT NULL,
    amount numeric(20,2) NOT NULL,
    transfer_method_id uuid NOT NULL,
    description text,
    group_expense_id uuid
);


ALTER TABLE public.debt_transactions OWNER TO postgres;

--
-- Name: friendship_requests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.friendship_requests (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    sender_profile_id uuid NOT NULL,
    recipient_profile_id uuid NOT NULL,
    blocked_at timestamp with time zone
);


ALTER TABLE public.friendship_requests OWNER TO postgres;

--
-- Name: friendships; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.friendships (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    profile_id1 uuid NOT NULL,
    profile_id2 uuid NOT NULL,
    type public.friendship_type NOT NULL,
    CONSTRAINT profile_order CHECK ((profile_id1 < profile_id2))
);


ALTER TABLE public.friendships OWNER TO postgres;

--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.goose_db_version OWNER TO postgres;

--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
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
-- Name: group_expense_bills; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_expense_bills (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    image_name text NOT NULL,
    group_expense_id uuid NOT NULL,
    status text DEFAULT 'PENDING'::text NOT NULL,
    extracted_text text
);


ALTER TABLE public.group_expense_bills OWNER TO postgres;

--
-- Name: group_expense_item_participants; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_expense_item_participants (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expense_item_id uuid NOT NULL,
    profile_id uuid NOT NULL,
    share numeric(20,4),
    weight smallint DEFAULT 1 NOT NULL,
    allocated_amount numeric(20,2) DEFAULT 0 NOT NULL
);


ALTER TABLE public.group_expense_item_participants OWNER TO postgres;

--
-- Name: group_expense_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_expense_items (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    group_expense_id uuid NOT NULL,
    name text NOT NULL,
    amount numeric(20,2) NOT NULL,
    quantity integer NOT NULL
);


ALTER TABLE public.group_expense_items OWNER TO postgres;

--
-- Name: group_expense_other_fee_participants; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_expense_other_fee_participants (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    other_fee_id uuid NOT NULL,
    profile_id uuid NOT NULL,
    share_amount numeric(20,2) NOT NULL
);


ALTER TABLE public.group_expense_other_fee_participants OWNER TO postgres;

--
-- Name: group_expense_other_fees; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_expense_other_fees (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    group_expense_id uuid NOT NULL,
    name text NOT NULL,
    amount numeric(20,2) NOT NULL,
    calculation_method public.fee_calculation_method
);


ALTER TABLE public.group_expense_other_fees OWNER TO postgres;

--
-- Name: group_expense_participants; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_expense_participants (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    group_expense_id uuid NOT NULL,
    participant_profile_id uuid NOT NULL,
    share_amount numeric(20,2) NOT NULL,
    proxy_profile_id uuid
);


ALTER TABLE public.group_expense_participants OWNER TO postgres;

--
-- Name: group_expenses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_expenses (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    description text,
    status text DEFAULT 'DRAFT'::text NOT NULL,
    total_amount numeric(20,2) NOT NULL,
    items_total numeric(20,2) DEFAULT 0 NOT NULL,
    fees_total numeric(20,2) DEFAULT 0 NOT NULL,
    payer_profile_id uuid,
    creator_profile_id uuid NOT NULL,
    processed boolean DEFAULT false NOT NULL
);


ALTER TABLE public.group_expenses OWNER TO postgres;

--
-- Name: notifications; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notifications (
    id uuid DEFAULT uuidv7() NOT NULL,
    profile_id uuid NOT NULL,
    type text NOT NULL,
    entity_type text NOT NULL,
    entity_id uuid NOT NULL,
    metadata jsonb,
    read_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    pushed_at timestamp with time zone
);


ALTER TABLE public.notifications OWNER TO postgres;

--
-- Name: oauth_accounts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.oauth_accounts (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    user_id uuid NOT NULL,
    provider text NOT NULL,
    provider_id text NOT NULL,
    email text
);


ALTER TABLE public.oauth_accounts OWNER TO postgres;

--
-- Name: password_reset_tokens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.password_reset_tokens (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    user_id uuid NOT NULL,
    token character varying(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL
);


ALTER TABLE public.password_reset_tokens OWNER TO postgres;

--
-- Name: plan_versions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.plan_versions (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    plan_id uuid NOT NULL,
    price_amount numeric(20,2) NOT NULL,
    price_currency character varying(3) NOT NULL,
    billing_interval text NOT NULL,
    bill_uploads_daily smallint NOT NULL,
    bill_uploads_monthly smallint NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    is_default boolean DEFAULT false NOT NULL
);


ALTER TABLE public.plan_versions OWNER TO postgres;

--
-- Name: plans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.plans (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT false NOT NULL,
    priority integer DEFAULT 1 NOT NULL
);


ALTER TABLE public.plans OWNER TO postgres;

--
-- Name: profile_transfer_methods; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.profile_transfer_methods (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    profile_id uuid NOT NULL,
    transfer_method_id uuid NOT NULL,
    account_name text NOT NULL,
    account_number text NOT NULL
);


ALTER TABLE public.profile_transfer_methods OWNER TO postgres;

--
-- Name: push_subscriptions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.push_subscriptions (
    id uuid DEFAULT uuidv7() NOT NULL,
    profile_id uuid NOT NULL,
    endpoint text NOT NULL,
    keys jsonb NOT NULL,
    user_agent text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    session_id uuid
);


ALTER TABLE public.push_subscriptions OWNER TO postgres;

--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.refresh_tokens (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    session_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL
);


ALTER TABLE public.refresh_tokens OWNER TO postgres;

--
-- Name: related_profiles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.related_profiles (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    real_profile_id uuid NOT NULL,
    anon_profile_id uuid NOT NULL
);


ALTER TABLE public.related_profiles OWNER TO postgres;

--
-- Name: sessions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sessions (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    user_id uuid NOT NULL,
    device_id text,
    last_used_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.sessions OWNER TO postgres;

--
-- Name: subscription_payments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subscription_payments (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    subscription_id uuid NOT NULL,
    amount numeric(20,2) NOT NULL,
    currency text NOT NULL,
    gateway text NOT NULL,
    gateway_transaction_id text,
    gateway_subscription_id text,
    status text NOT NULL,
    failure_reason text,
    starts_at timestamp with time zone,
    ends_at timestamp with time zone,
    gateway_event_id text,
    paid_at timestamp with time zone,
    expired_at timestamp with time zone
);


ALTER TABLE public.subscription_payments OWNER TO postgres;

--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subscriptions (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    profile_id uuid NOT NULL,
    plan_version_id uuid NOT NULL,
    ends_at timestamp with time zone,
    canceled_at timestamp with time zone,
    auto_renew boolean DEFAULT true NOT NULL,
    status text NOT NULL,
    current_period_start timestamp with time zone,
    current_period_end timestamp with time zone
);


ALTER TABLE public.subscriptions OWNER TO postgres;

--
-- Name: transfer_methods; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.transfer_methods (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    name text NOT NULL,
    display text NOT NULL,
    icon_url text,
    parent_id uuid,
    count bigint DEFAULT 0
);


ALTER TABLE public.transfer_methods OWNER TO postgres;

--
-- Name: user_profiles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_profiles (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    user_id uuid,
    name text NOT NULL,
    avatar text
);


ALTER TABLE public.user_profiles OWNER TO postgres;

--
-- Name: COLUMN user_profiles.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.user_profiles.user_id IS 'Nullable. Can be NULL for peers who do not have an account in the app';


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    verified_at timestamp with time zone
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: admin_roles admin_roles_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_roles
    ADD CONSTRAINT admin_roles_name_key UNIQUE (name);


--
-- Name: admin_roles admin_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_roles
    ADD CONSTRAINT admin_roles_pkey PRIMARY KEY (id);


--
-- Name: admin_users admin_users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_users
    ADD CONSTRAINT admin_users_email_key UNIQUE (email);


--
-- Name: admin_users admin_users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_users
    ADD CONSTRAINT admin_users_pkey PRIMARY KEY (id);


--
-- Name: admin_users_roles admin_users_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_users_roles
    ADD CONSTRAINT admin_users_roles_pkey PRIMARY KEY (id);


--
-- Name: debt_transactions debt_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.debt_transactions
    ADD CONSTRAINT debt_transactions_pkey PRIMARY KEY (id);


--
-- Name: friendship_requests friendship_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.friendship_requests
    ADD CONSTRAINT friendship_requests_pkey PRIMARY KEY (id);


--
-- Name: friendships friendships_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.friendships
    ADD CONSTRAINT friendships_pkey PRIMARY KEY (id);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: group_expense_bills group_expense_bills_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_bills
    ADD CONSTRAINT group_expense_bills_pkey PRIMARY KEY (id);


--
-- Name: group_expense_item_participants group_expense_item_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_item_participants
    ADD CONSTRAINT group_expense_item_participants_pkey PRIMARY KEY (id);


--
-- Name: group_expense_items group_expense_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_items
    ADD CONSTRAINT group_expense_items_pkey PRIMARY KEY (id);


--
-- Name: group_expense_other_fee_participants group_expense_other_fee_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_other_fee_participants
    ADD CONSTRAINT group_expense_other_fee_participants_pkey PRIMARY KEY (id);


--
-- Name: group_expense_other_fees group_expense_other_fees_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_other_fees
    ADD CONSTRAINT group_expense_other_fees_pkey PRIMARY KEY (id);


--
-- Name: group_expense_participants group_expense_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_participants
    ADD CONSTRAINT group_expense_participants_pkey PRIMARY KEY (id);


--
-- Name: group_expenses group_expenses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expenses
    ADD CONSTRAINT group_expenses_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: oauth_accounts oauth_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.oauth_accounts
    ADD CONSTRAINT oauth_accounts_pkey PRIMARY KEY (id);


--
-- Name: oauth_accounts oauth_accounts_provider_provider_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.oauth_accounts
    ADD CONSTRAINT oauth_accounts_provider_provider_id_unique UNIQUE (provider, provider_id);


--
-- Name: password_reset_tokens password_reset_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (id);


--
-- Name: password_reset_tokens password_reset_tokens_token_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_token_key UNIQUE (token);


--
-- Name: plan_versions plan_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.plan_versions
    ADD CONSTRAINT plan_versions_pkey PRIMARY KEY (id);


--
-- Name: plans plans_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.plans
    ADD CONSTRAINT plans_pkey PRIMARY KEY (id);


--
-- Name: profile_transfer_methods profile_transfer_methods_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile_transfer_methods
    ADD CONSTRAINT profile_transfer_methods_pkey PRIMARY KEY (id);


--
-- Name: push_subscriptions push_subscriptions_endpoint_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.push_subscriptions
    ADD CONSTRAINT push_subscriptions_endpoint_key UNIQUE (endpoint);


--
-- Name: push_subscriptions push_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.push_subscriptions
    ADD CONSTRAINT push_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: related_profiles related_profiles_anon_profile_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.related_profiles
    ADD CONSTRAINT related_profiles_anon_profile_id_key UNIQUE (anon_profile_id);


--
-- Name: related_profiles related_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.related_profiles
    ADD CONSTRAINT related_profiles_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: subscription_payments subscription_payments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscription_payments
    ADD CONSTRAINT subscription_payments_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: transfer_methods transfer_methods_name_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transfer_methods
    ADD CONSTRAINT transfer_methods_name_unique UNIQUE (name);


--
-- Name: transfer_methods transfer_methods_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transfer_methods
    ADD CONSTRAINT transfer_methods_pkey PRIMARY KEY (id);


--
-- Name: group_expense_item_participants unique_expense_item_profile; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_item_participants
    ADD CONSTRAINT unique_expense_item_profile UNIQUE (expense_item_id, profile_id);


--
-- Name: group_expense_participants unique_expense_profile; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_participants
    ADD CONSTRAINT unique_expense_profile UNIQUE (group_expense_id, participant_profile_id);


--
-- Name: group_expense_other_fee_participants unique_fee_participant; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_other_fee_participants
    ADD CONSTRAINT unique_fee_participant UNIQUE (other_fee_id, profile_id);


--
-- Name: friendships unique_friendship; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.friendships
    ADD CONSTRAINT unique_friendship UNIQUE (profile_id1, profile_id2);


--
-- Name: user_profiles user_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: admin_users_roles_role_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX admin_users_roles_role_id_idx ON public.admin_users_roles USING btree (role_id);


--
-- Name: admin_users_roles_unique_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX admin_users_roles_unique_idx ON public.admin_users_roles USING btree (user_id, role_id);


--
-- Name: admin_users_roles_user_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX admin_users_roles_user_id_idx ON public.admin_users_roles USING btree (user_id);


--
-- Name: debt_transactions_borrower_profile_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX debt_transactions_borrower_profile_id_idx ON public.debt_transactions USING btree (borrower_profile_id);


--
-- Name: debt_transactions_created_at_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX debt_transactions_created_at_idx ON public.debt_transactions USING btree (created_at);


--
-- Name: debt_transactions_group_expense_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX debt_transactions_group_expense_id_idx ON public.debt_transactions USING btree (group_expense_id);


--
-- Name: debt_transactions_lender_profile_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX debt_transactions_lender_profile_id_idx ON public.debt_transactions USING btree (lender_profile_id);


--
-- Name: debt_transactions_transfer_method_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX debt_transactions_transfer_method_id_idx ON public.debt_transactions USING btree (transfer_method_id);


--
-- Name: friendships_profile_id1_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX friendships_profile_id1_idx ON public.friendships USING btree (profile_id1);


--
-- Name: friendships_profile_id2_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX friendships_profile_id2_idx ON public.friendships USING btree (profile_id2);


--
-- Name: friendships_type_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX friendships_type_idx ON public.friendships USING btree (type);


--
-- Name: group_expense_bills_created_at_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX group_expense_bills_created_at_idx ON public.group_expense_bills USING btree (created_at);


--
-- Name: group_expense_participants_created_at_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX group_expense_participants_created_at_idx ON public.group_expense_participants USING btree (created_at);


--
-- Name: group_expense_participants_group_expense_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX group_expense_participants_group_expense_id_idx ON public.group_expense_participants USING btree (group_expense_id);


--
-- Name: group_expense_participants_participant_profile_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX group_expense_participants_participant_profile_id_idx ON public.group_expense_participants USING btree (participant_profile_id);


--
-- Name: group_expenses_created_at_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX group_expenses_created_at_idx ON public.group_expenses USING btree (created_at);


--
-- Name: group_expenses_payer_profile_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX group_expenses_payer_profile_id_idx ON public.group_expenses USING btree (payer_profile_id);


--
-- Name: idx_password_reset_tokens_expires_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_reset_tokens_expires_at ON public.password_reset_tokens USING btree (expires_at);


--
-- Name: idx_password_reset_tokens_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_reset_tokens_user_id ON public.password_reset_tokens USING btree (user_id);


--
-- Name: idx_session_tokens; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_session_tokens ON public.refresh_tokens USING btree (session_id);


--
-- Name: idx_token_expiry; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_token_expiry ON public.refresh_tokens USING btree (expires_at);


--
-- Name: idx_token_lookup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_token_lookup ON public.refresh_tokens USING btree (token_hash);


--
-- Name: idx_user_profiles_name_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_profiles_name_trgm ON public.user_profiles USING gin (name public.gin_trgm_ops);


--
-- Name: notifications_profile_unread_created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX notifications_profile_unread_created_idx ON public.notifications USING btree (profile_id, created_at DESC) WHERE (read_at IS NULL);


--
-- Name: notifications_unique_entity_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX notifications_unique_entity_idx ON public.notifications USING btree (profile_id, type, entity_type, entity_id);


--
-- Name: plan_versions_plan_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX plan_versions_plan_id_idx ON public.plan_versions USING btree (plan_id);


--
-- Name: profile_transfer_methods_profile_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX profile_transfer_methods_profile_id_idx ON public.profile_transfer_methods USING btree (profile_id);


--
-- Name: profile_transfer_methods_transfer_method_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX profile_transfer_methods_transfer_method_id_idx ON public.profile_transfer_methods USING btree (transfer_method_id);


--
-- Name: push_subscriptions_profile_endpoint_unique_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX push_subscriptions_profile_endpoint_unique_idx ON public.push_subscriptions USING btree (profile_id, endpoint);


--
-- Name: push_subscriptions_session_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX push_subscriptions_session_id_idx ON public.push_subscriptions USING btree (session_id);


--
-- Name: sessions_user_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX sessions_user_id_idx ON public.sessions USING btree (user_id);


--
-- Name: subscription_payments_gateway_unique_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX subscription_payments_gateway_unique_idx ON public.subscription_payments USING btree (gateway_transaction_id, gateway);


--
-- Name: subscription_payments_subscription_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX subscription_payments_subscription_id_idx ON public.subscription_payments USING btree (subscription_id);


--
-- Name: subscriptions_plan_version_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX subscriptions_plan_version_id_idx ON public.subscriptions USING btree (plan_version_id);


--
-- Name: subscriptions_profile_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX subscriptions_profile_id_idx ON public.subscriptions USING btree (profile_id);


--
-- Name: transfer_methods_parent_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX transfer_methods_parent_id_idx ON public.transfer_methods USING btree (parent_id) WHERE (parent_id IS NOT NULL);


--
-- Name: unique_incomplete_payment_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX unique_incomplete_payment_idx ON public.subscription_payments USING btree (subscription_id) WHERE (status = ANY (ARRAY['pending'::text, 'processing'::text]));


--
-- Name: user_profiles_name_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX user_profiles_name_idx ON public.user_profiles USING btree (name);


--
-- Name: user_profiles_user_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX user_profiles_user_id_idx ON public.user_profiles USING btree (user_id);


--
-- Name: admin_users_roles admin_users_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_users_roles
    ADD CONSTRAINT admin_users_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.admin_roles(id) ON DELETE CASCADE;


--
-- Name: admin_users_roles admin_users_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_users_roles
    ADD CONSTRAINT admin_users_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.admin_users(id) ON DELETE CASCADE;


--
-- Name: debt_transactions debt_transactions_borrower_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.debt_transactions
    ADD CONSTRAINT debt_transactions_borrower_profile_id_fkey FOREIGN KEY (borrower_profile_id) REFERENCES public.user_profiles(id);


--
-- Name: debt_transactions debt_transactions_lender_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.debt_transactions
    ADD CONSTRAINT debt_transactions_lender_profile_id_fkey FOREIGN KEY (lender_profile_id) REFERENCES public.user_profiles(id);


--
-- Name: debt_transactions debt_transactions_transfer_method_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.debt_transactions
    ADD CONSTRAINT debt_transactions_transfer_method_id_fkey FOREIGN KEY (transfer_method_id) REFERENCES public.transfer_methods(id);


--
-- Name: friendship_requests friendship_requests_recipient_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.friendship_requests
    ADD CONSTRAINT friendship_requests_recipient_profile_id_fkey FOREIGN KEY (recipient_profile_id) REFERENCES public.user_profiles(id);


--
-- Name: friendship_requests friendship_requests_sender_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.friendship_requests
    ADD CONSTRAINT friendship_requests_sender_profile_id_fkey FOREIGN KEY (sender_profile_id) REFERENCES public.user_profiles(id);


--
-- Name: friendships friendships_profile_id1_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.friendships
    ADD CONSTRAINT friendships_profile_id1_fkey FOREIGN KEY (profile_id1) REFERENCES public.user_profiles(id);


--
-- Name: friendships friendships_profile_id2_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.friendships
    ADD CONSTRAINT friendships_profile_id2_fkey FOREIGN KEY (profile_id2) REFERENCES public.user_profiles(id);


--
-- Name: group_expense_bills group_expense_bills_group_expense_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_bills
    ADD CONSTRAINT group_expense_bills_group_expense_id_fkey FOREIGN KEY (group_expense_id) REFERENCES public.group_expenses(id) ON DELETE CASCADE;


--
-- Name: group_expense_item_participants group_expense_item_participants_expense_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_item_participants
    ADD CONSTRAINT group_expense_item_participants_expense_item_id_fkey FOREIGN KEY (expense_item_id) REFERENCES public.group_expense_items(id) ON DELETE CASCADE;


--
-- Name: group_expense_item_participants group_expense_item_participants_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_item_participants
    ADD CONSTRAINT group_expense_item_participants_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.user_profiles(id);


--
-- Name: group_expense_items group_expense_items_group_expense_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_items
    ADD CONSTRAINT group_expense_items_group_expense_id_fkey FOREIGN KEY (group_expense_id) REFERENCES public.group_expenses(id) ON DELETE CASCADE;


--
-- Name: group_expense_other_fee_participants group_expense_other_fee_participants_other_fee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_other_fee_participants
    ADD CONSTRAINT group_expense_other_fee_participants_other_fee_id_fkey FOREIGN KEY (other_fee_id) REFERENCES public.group_expense_other_fees(id);


--
-- Name: group_expense_other_fee_participants group_expense_other_fee_participants_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_other_fee_participants
    ADD CONSTRAINT group_expense_other_fee_participants_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.user_profiles(id);


--
-- Name: group_expense_other_fees group_expense_other_fees_group_expense_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_other_fees
    ADD CONSTRAINT group_expense_other_fees_group_expense_id_fkey FOREIGN KEY (group_expense_id) REFERENCES public.group_expenses(id) ON DELETE CASCADE;


--
-- Name: group_expense_participants group_expense_participants_group_expense_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_participants
    ADD CONSTRAINT group_expense_participants_group_expense_id_fkey FOREIGN KEY (group_expense_id) REFERENCES public.group_expenses(id) ON DELETE CASCADE;


--
-- Name: group_expense_participants group_expense_participants_participant_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_participants
    ADD CONSTRAINT group_expense_participants_participant_profile_id_fkey FOREIGN KEY (participant_profile_id) REFERENCES public.user_profiles(id);


--
-- Name: group_expense_participants group_expense_participants_proxy_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expense_participants
    ADD CONSTRAINT group_expense_participants_proxy_profile_id_fkey FOREIGN KEY (proxy_profile_id) REFERENCES public.user_profiles(id);


--
-- Name: group_expenses group_expenses_creator_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expenses
    ADD CONSTRAINT group_expenses_creator_profile_id_fkey FOREIGN KEY (creator_profile_id) REFERENCES public.user_profiles(id);


--
-- Name: group_expenses group_expenses_payer_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_expenses
    ADD CONSTRAINT group_expenses_payer_profile_id_fkey FOREIGN KEY (payer_profile_id) REFERENCES public.user_profiles(id);


--
-- Name: notifications notifications_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.user_profiles(id);


--
-- Name: oauth_accounts oauth_accounts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.oauth_accounts
    ADD CONSTRAINT oauth_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: password_reset_tokens password_reset_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: plan_versions plan_versions_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.plan_versions
    ADD CONSTRAINT plan_versions_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.plans(id);


--
-- Name: profile_transfer_methods profile_transfer_methods_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile_transfer_methods
    ADD CONSTRAINT profile_transfer_methods_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.user_profiles(id);


--
-- Name: profile_transfer_methods profile_transfer_methods_transfer_method_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile_transfer_methods
    ADD CONSTRAINT profile_transfer_methods_transfer_method_id_fkey FOREIGN KEY (transfer_method_id) REFERENCES public.transfer_methods(id);


--
-- Name: push_subscriptions push_subscriptions_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.push_subscriptions
    ADD CONSTRAINT push_subscriptions_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.user_profiles(id);


--
-- Name: push_subscriptions push_subscriptions_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.push_subscriptions
    ADD CONSTRAINT push_subscriptions_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.sessions(id) ON DELETE CASCADE;


--
-- Name: refresh_tokens refresh_tokens_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.sessions(id) ON DELETE CASCADE;


--
-- Name: related_profiles related_profiles_anon_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.related_profiles
    ADD CONSTRAINT related_profiles_anon_profile_id_fkey FOREIGN KEY (anon_profile_id) REFERENCES public.user_profiles(id);


--
-- Name: related_profiles related_profiles_real_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.related_profiles
    ADD CONSTRAINT related_profiles_real_profile_id_fkey FOREIGN KEY (real_profile_id) REFERENCES public.user_profiles(id);


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: subscription_payments subscription_payments_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscription_payments
    ADD CONSTRAINT subscription_payments_subscription_id_fkey FOREIGN KEY (subscription_id) REFERENCES public.subscriptions(id);


--
-- Name: subscriptions subscriptions_plan_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_plan_version_id_fkey FOREIGN KEY (plan_version_id) REFERENCES public.plan_versions(id);


--
-- Name: subscriptions subscriptions_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.user_profiles(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict l53tlQ1rF2r6rqVgGl707x5be7yFSnrHUzLlRQ72v1pbmgWEhUdMSdg0TF1irJD

