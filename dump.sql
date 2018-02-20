--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.5
-- Dumped by pg_dump version 10.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

--
-- Name: message_status; Type: TYPE; Schema: public; Owner: dappmanager
--

CREATE TYPE message_status AS ENUM (
);


ALTER TYPE message_status OWNER TO dappmanager;

--
-- Name: message_type; Type: TYPE; Schema: public; Owner: dappmanager
--

CREATE TYPE message_type AS ENUM (
);


ALTER TYPE message_type OWNER TO dappmanager;

--
-- Name: privatix_tokens; Type: DOMAIN; Schema: public; Owner: dappmanager
--

CREATE DOMAIN privatix_tokens AS numeric(64,8);


ALTER DOMAIN privatix_tokens OWNER TO dappmanager;

--
-- Name: service_type; Type: TYPE; Schema: public; Owner: dappmanager
--

CREATE TYPE service_type AS ENUM (
    'vpn'
);


ALTER TYPE service_type OWNER TO dappmanager;

--
-- Name: state_channels_state; Type: TYPE; Schema: public; Owner: dappmanager
--

CREATE TYPE state_channels_state AS ENUM (
    'open',
    'cooperative_closed',
    'uncooperative_closed',
    'waiting_for_client_password'
);


ALTER TYPE state_channels_state OWNER TO dappmanager;

--
-- Name: vpn_protocol; Type: TYPE; Schema: public; Owner: dappmanager
--

CREATE TYPE vpn_protocol AS ENUM (
    'tcp',
    'udp'
);


ALTER TYPE vpn_protocol OWNER TO dappmanager;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: agents; Type: TABLE; Schema: public; Owner: dappmanager
--

CREATE TABLE agents (
    id bigint NOT NULL,
    public_key bytea NOT NULL,
    private_key bytea NOT NULL
);


ALTER TABLE agents OWNER TO dappmanager;

--
-- Name: agents_id_seq; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE agents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE agents_id_seq OWNER TO dappmanager;

--
-- Name: agents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE agents_id_seq OWNED BY agents.id;


--
-- Name: clients; Type: TABLE; Schema: public; Owner: dappmanager
--

CREATE TABLE clients (
    id bigint NOT NULL,
    public_key bytea NOT NULL,
    private_key bytea NOT NULL
);


ALTER TABLE clients OWNER TO dappmanager;

--
-- Name: clients_id_seq; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE clients_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE clients_id_seq OWNER TO dappmanager;

--
-- Name: clients_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE clients_id_seq OWNED BY clients.id;


--
-- Name: messages; Type: TABLE; Schema: public; Owner: dappmanager
--

CREATE TABLE messages (
    id bigint NOT NULL,
    type message_type,
    raw text,
    status message_status NOT NULL
);


ALTER TABLE messages OWNER TO dappmanager;

--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE messages_id_seq OWNER TO dappmanager;

--
-- Name: messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE messages_id_seq OWNED BY messages.id;


--
-- Name: offerings; Type: TABLE; Schema: public; Owner: dappmanager
--

CREATE TABLE offerings (
    id bigint NOT NULL,
    hash_hex text NOT NULL,
    agent_public_key bytea NOT NULL,
    service_type service_type DEFAULT 'vpn'::service_type NOT NULL,
    uuid uuid NOT NULL,
    signature_hex text NOT NULL
);


ALTER TABLE offerings OWNER TO dappmanager;

--
-- Name: offerings_id_seq; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE offerings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE offerings_id_seq OWNER TO dappmanager;

--
-- Name: offerings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE offerings_id_seq OWNED BY offerings.id;


--
-- Name: offerings_vpn; Type: TABLE; Schema: public; Owner: dappmanager
--

CREATE TABLE offerings_vpn (
    offering_id bigint NOT NULL,
    hash_hex text NOT NULL,
    country text NOT NULL,
    service_supply text NOT NULL,
    unit_price privatix_tokens NOT NULL,
    min_units bigint NOT NULL,
    max_units bigint NOT NULL,
    billing_interval bigint NOT NULL,
    billing_interval_time_deviation bigint NOT NULL,
    free_intervals bigint NOT NULL,
    min_download_bps bigint,
    protocol vpn_protocol DEFAULT 'tcp'::vpn_protocol,
    id bigint NOT NULL,
    template_version integer DEFAULT 1 NOT NULL
);


ALTER TABLE offerings_vpn OWNER TO dappmanager;

--
-- Name: offerings_vpn_id_seq; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE offerings_vpn_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE offerings_vpn_id_seq OWNER TO dappmanager;

--
-- Name: offerings_vpn_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE offerings_vpn_id_seq OWNED BY offerings_vpn.offering_id;


--
-- Name: offerings_vpn_id_seq1; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE offerings_vpn_id_seq1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE offerings_vpn_id_seq1 OWNER TO dappmanager;

--
-- Name: offerings_vpn_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE offerings_vpn_id_seq1 OWNED BY offerings_vpn.id;


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: dappmanager
--

CREATE TABLE sessions (
    id bigint NOT NULL,
    state_channel_id bigint NOT NULL,
    started timestamp with time zone NOT NULL,
    stopped timestamp with time zone NOT NULL,
    outgoing_bytes bigint DEFAULT 0 NOT NULL,
    incoming_bytes bigint DEFAULT 0 NOT NULL
);


ALTER TABLE sessions OWNER TO dappmanager;

--
-- Name: sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sessions_id_seq OWNER TO dappmanager;

--
-- Name: sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE sessions_id_seq OWNED BY sessions.id;


--
-- Name: sessions_incoming_bytes_seq; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE sessions_incoming_bytes_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sessions_incoming_bytes_seq OWNER TO dappmanager;

--
-- Name: sessions_incoming_bytes_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE sessions_incoming_bytes_seq OWNED BY sessions.incoming_bytes;


--
-- Name: sessions_outgoing_bytes_seq; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE sessions_outgoing_bytes_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sessions_outgoing_bytes_seq OWNER TO dappmanager;

--
-- Name: sessions_outgoing_bytes_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE sessions_outgoing_bytes_seq OWNED BY sessions.outgoing_bytes;


--
-- Name: sessions_state_channel_id_seq; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE sessions_state_channel_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sessions_state_channel_id_seq OWNER TO dappmanager;

--
-- Name: sessions_state_channel_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE sessions_state_channel_id_seq OWNED BY sessions.state_channel_id;


--
-- Name: settings; Type: TABLE; Schema: public; Owner: dappmanager
--

CREATE TABLE settings (
    key text NOT NULL,
    value text NOT NULL,
    description text
);


ALTER TABLE settings OWNER TO dappmanager;

--
-- Name: state_channels; Type: TABLE; Schema: public; Owner: dappmanager
--

CREATE TABLE state_channels (
    id bigint NOT NULL,
    agent_id bigint NOT NULL,
    client_id bigint NOT NULL,
    offering_id bigint NOT NULL,
    block bigint NOT NULL,
    total_deposit privatix_tokens NOT NULL,
    state state_channels_state NOT NULL,
    closed_deposit privatix_tokens DEFAULT 0 NOT NULL,
    password text NOT NULL,
    salt text NOT NULL,
    last_receipt_balance privatix_tokens DEFAULT 0 NOT NULL,
    last_receipt_balance_msg_siig text
);


ALTER TABLE state_channels OWNER TO dappmanager;

--
-- Name: state_channels_id_seq; Type: SEQUENCE; Schema: public; Owner: dappmanager
--

CREATE SEQUENCE state_channels_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE state_channels_id_seq OWNER TO dappmanager;

--
-- Name: state_channels_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dappmanager
--

ALTER SEQUENCE state_channels_id_seq OWNED BY state_channels.id;


--
-- Name: agents id; Type: DEFAULT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY agents ALTER COLUMN id SET DEFAULT nextval('agents_id_seq'::regclass);


--
-- Name: clients id; Type: DEFAULT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY clients ALTER COLUMN id SET DEFAULT nextval('clients_id_seq'::regclass);


--
-- Name: messages id; Type: DEFAULT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY messages ALTER COLUMN id SET DEFAULT nextval('messages_id_seq'::regclass);


--
-- Name: offerings id; Type: DEFAULT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY offerings ALTER COLUMN id SET DEFAULT nextval('offerings_id_seq'::regclass);


--
-- Name: offerings_vpn id; Type: DEFAULT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY offerings_vpn ALTER COLUMN id SET DEFAULT nextval('offerings_vpn_id_seq1'::regclass);


--
-- Name: sessions id; Type: DEFAULT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY sessions ALTER COLUMN id SET DEFAULT nextval('sessions_id_seq'::regclass);


--
-- Name: sessions state_channel_id; Type: DEFAULT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY sessions ALTER COLUMN state_channel_id SET DEFAULT nextval('sessions_state_channel_id_seq'::regclass);


--
-- Name: state_channels id; Type: DEFAULT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY state_channels ALTER COLUMN id SET DEFAULT nextval('state_channels_id_seq'::regclass);


--
-- Name: agents agents_pkey; Type: CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY agents
    ADD CONSTRAINT agents_pkey PRIMARY KEY (id);


--
-- Name: clients clients_pkey; Type: CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY clients
    ADD CONSTRAINT clients_pkey PRIMARY KEY (id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: offerings offerings_pkey; Type: CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY offerings
    ADD CONSTRAINT offerings_pkey PRIMARY KEY (id);


--
-- Name: offerings_vpn offerings_vpn_id_pk; Type: CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY offerings_vpn
    ADD CONSTRAINT offerings_vpn_id_pk PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: state_channels state_channels_pkey; Type: CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY state_channels
    ADD CONSTRAINT state_channels_pkey PRIMARY KEY (id);


--
-- Name: agents_id_uindex; Type: INDEX; Schema: public; Owner: dappmanager
--

CREATE UNIQUE INDEX agents_id_uindex ON agents USING btree (id);


--
-- Name: clients_id_uindex; Type: INDEX; Schema: public; Owner: dappmanager
--

CREATE UNIQUE INDEX clients_id_uindex ON clients USING btree (id);


--
-- Name: messages_id_uindex; Type: INDEX; Schema: public; Owner: dappmanager
--

CREATE UNIQUE INDEX messages_id_uindex ON messages USING btree (id);


--
-- Name: offerings_id_uindex; Type: INDEX; Schema: public; Owner: dappmanager
--

CREATE UNIQUE INDEX offerings_id_uindex ON offerings USING btree (id);


--
-- Name: offerings_vpn_hash_hex_uindex; Type: INDEX; Schema: public; Owner: dappmanager
--

CREATE UNIQUE INDEX offerings_vpn_hash_hex_uindex ON offerings_vpn USING btree (hash_hex);


--
-- Name: offerings_vpn_id_uindex; Type: INDEX; Schema: public; Owner: dappmanager
--

CREATE UNIQUE INDEX offerings_vpn_id_uindex ON offerings_vpn USING btree (offering_id);


--
-- Name: settings_key_uindex; Type: INDEX; Schema: public; Owner: dappmanager
--

CREATE UNIQUE INDEX settings_key_uindex ON settings USING btree (key);


--
-- Name: state_channels_id_uindex; Type: INDEX; Schema: public; Owner: dappmanager
--

CREATE UNIQUE INDEX state_channels_id_uindex ON state_channels USING btree (id);


--
-- Name: offerings_vpn offerings_vpn_offerings_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY offerings_vpn
    ADD CONSTRAINT offerings_vpn_offerings_id_fk FOREIGN KEY (offering_id) REFERENCES offerings(id);


--
-- Name: sessions sessions_state_channels_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY sessions
    ADD CONSTRAINT sessions_state_channels_id_fk FOREIGN KEY (state_channel_id) REFERENCES state_channels(id);


--
-- Name: state_channels state_channels_agents_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY state_channels
    ADD CONSTRAINT state_channels_agents_id_fk FOREIGN KEY (agent_id) REFERENCES agents(id);


--
-- Name: state_channels state_channels_clients_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: dappmanager
--

ALTER TABLE ONLY state_channels
    ADD CONSTRAINT state_channels_clients_id_fk FOREIGN KEY (client_id) REFERENCES clients(id);


--
-- PostgreSQL database dump complete
--

