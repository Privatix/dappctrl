BEGIN TRANSACTION;

-- Service Units usage reporting type. Can be incremental or total. Indicates how reporting server will report usage of units.
DROP TYPE IF EXISTS usage_rep_type CASCADE;
CREATE TYPE usage_rep_type AS ENUM ('incremental', 'total');

-- Templates kinds.
DROP TYPE IF EXISTS tpl_kind CASCADE;
CREATE TYPE tpl_kind AS ENUM ('offer', 'auth', 'access');

-- Billing types.
DROP TYPE IF EXISTS bill_type CASCADE;
CREATE TYPE bill_type AS ENUM ('prepaid','postpaid');

-- Unit types. Used for billing calculation.
DROP TYPE IF EXISTS unit_type CASCADE;
CREATE TYPE unit_type AS ENUM ('units','seconds');

-- Contract types.
DROP TYPE IF EXISTS contract_type CASCADE;
CREATE TYPE contract_type AS ENUM ('ptc','psc');


-- SHA3-256 in base64 (RFC-4648).
DROP DOMAIN IF EXISTS sha3_256 CASCADE;
CREATE DOMAIN sha3_256 AS char(44);

-- Etehereum address
DROP DOMAIN IF EXISTS eth_addr CASCADE;
CREATE DOMAIN eth_addr AS char(28);

-- Service operational status.
DROP TYPE IF EXISTS svc_status CASCADE;
CREATE TYPE svc_status AS ENUM (
    'pending', -- Service is still not fully setup and cannot be used. E.g. waiting for authentication message/endpoint message.
    'active', -- service is now active and can be used.
    'suspended', -- service usage is not allowed. Usually used to temporary disallow access.
    'terminated' -- service is permanently deactivated.
);

-- State channel states.
DROP TYPE IF EXISTS chan_status CASCADE;
CREATE TYPE chan_status AS ENUM (
    'pending', -- waiting to be opened
    'active', -- opened
    'wait_coop', -- waiting to be closed cooperatively
    'closed_coop', -- closed cooperatively
    'wait_challenge', -- waiting to start challenge period
    'in_challenge', -- challenge period started for uncooperative close
    'wait_uncoop', -- waiting for settling state channel uncooperatively
    'closed_uncoop' -- closed uncooperatively
);

-- Messages statuses.
DROP TYPE IF EXISTS msg_status CASCADE;
CREATE TYPE msg_status AS ENUM (
    'unpublished', -- saved in DB, but not published
    'bchain_publishing', -- publishing in blockchain
    'bchain_published', -- published in blockchain
    'msg_channel_published' -- published in messaging channel
);

-- Offering status
DROP TYPE IF EXISTS offer_status CASCADE;
CREATE TYPE offer_status AS ENUM (
    'empty', -- saved in DB, but not published to blockchain
    'register', -- in registration or registered in blockchain
    'remove' -- being removed or already removed from blockchain
);

-- Transaction statuses.
DROP TYPE IF EXISTS tx_status CASCADE;
CREATE TYPE tx_status AS ENUM (
    'unsent', -- saved in DB, but not sent
    'sent', -- sent w/o error to eth node
    'mined', -- tx mined
    'uncle' -- tx is went to uncle block
);

-- Job creator.
DROP TYPE IF EXISTS job_creator CASCADE;
CREATE TYPE job_creator AS ENUM (
    'user', -- by user through UI
    'billing_checker', -- by billing checker procedure
    'bc_monitor', -- by blockchain monitor
    'task' -- by another task
);

-- Job status.
DROP TYPE IF EXISTS job_status CASCADE;
CREATE TYPE job_status AS ENUM (
    'active', -- processing or to be processed
    'done', -- successfully finished
    'failed', -- failed: retry limit is reached
    'canceled' -- canceled
);

-- Job related object types.
DROP TYPE IF EXISTS related_type CASCADE;
CREATE TYPE related_type AS ENUM (
    'offering', -- service offering
    'channel', -- state channel
    'endpoint' -- service endpoint
);

DROP TABLE IF EXISTS settings CASCADE;
CREATE TABLE settings (
    key text PRIMARY KEY,
    value text NOT NULL,
    description text
);

-- Accounts are ethereum accounts.
-- Accounts used to perform Client and/or Agent operations.
DROP TABLE IF EXISTS accounts CASCADE;
CREATE TABLE accounts (
    id uuid PRIMARY KEY,
    eth_addr eth_addr NOT NULL, -- ethereum address
    public_key text NOT NULL,
    private_key text NOT NULL,
    is_default boolean NOT NULL DEFAULT FALSE, -- default account
    in_use boolean NOT NULL DEFAULT TRUE -- this account is in use or not
);

-- Users are external party in distributed trade.
-- Each of them can play an agent role, a client role, or both of them.
DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users (
    id uuid PRIMARY KEY,
    eth_addr eth_addr NOT NULL, -- ethereum address
    public_key text NOT NULL
);

-- Templates.
DROP TABLE IF EXISTS templates CASCADE;
CREATE TABLE templates (
    id uuid PRIMARY KEY,
    hash sha3_256 NOT NULL,
    raw json NOT NULL,
    kind tpl_kind NOT NULL
);

-- Products. Used to store billing and action related settings.
DROP TABLE IF EXISTS products CASCADE;
CREATE TABLE products (
    id uuid PRIMARY KEY,
    name varchar(64) NOT NULL,
    offer_tpl_id uuid REFERENCES templates(id), -- enables product specific billing and actions support for Client
    -- offer_auth_id uuid REFERENCES templates(id), -- currently not in use. for future use.
    offer_access_id uuid REFERENCES templates(id), -- allows to identify endpoint message relation
    usage_rep_type usage_rep_type NOT NULL, -- for billing logic. Reporter provides increment or total usage
    is_server boolean NOT NULL -- product is defined as server (Agent) or client (Client)
);

-- Service offerings.
DROP TABLE IF EXISTS offerings CASCADE;
CREATE TABLE offerings (
    id uuid PRIMARY KEY,
    is_local boolean NOT NULL, -- created locally (by this Agent) or retreived (by this Client)
    tpl uuid NOT NULL REFERENCES templates(id), -- corresponding template
    product uuid NOT NULL REFERENCES products(id), -- enables product specific billing and actions support for Agent
    hash sha3_256 NOT NULL, -- offering hash
    status msg_status NOT NULL, -- message status
    offer_status offer_status NOT NULL, -- offer status in blockchain
    block_number_updated bigint
        CONSTRAINT positive_block_number_updated CHECK (offerings.block_number_updated > 0), -- block number, when offering was updated

    agent eth_addr NOT NULL,
    signature text NOT NULL, -- agent's signature
    service_name varchar(64) NOT NULL, -- name of service
    description text, -- description for UI
    country char(2) NOT NULL, -- ISO 3166-1 alpha-2
    supply int NOT NULL, -- maximum identical offerings for concurrent use through different state channels
    unit_name varchar(10) NOT NULL, -- like megabytes, minutes, etc
    unit_type unit_type NOT NULL, -- type of unit. Time or material.
    billing_type bill_type NOT NULL, -- prepaid/postpaid
    setup_price bigint NOT NULL, -- setup fee
    unit_price bigint NOT NULL,
    min_units bigint NOT NULL -- used to calculate min required deposit
        CONSTRAINT positive_min_units CHECK (offerings.min_units >= 0),

    max_unit bigint -- optional. If specified automatic termination can be invoked
        CONSTRAINT positive_max_unit CHECK (offerings.max_unit >= 0),

    billing_interval int NOT NULL -- every unit numbers, that should be paid, after free units consumed
        CONSTRAINT positive_billing_interval CHECK (offerings.billing_interval > 0),

    max_billing_unit_lag int NOT NULL --maximum tolerance for payment lag (in units)
        CONSTRAINT positive_max_billing_unit_lag CHECK (offerings.max_billing_unit_lag >= 0),

    max_suspended_time int NOT NULL -- maximum time in suspend state, after which service will be terminated (in seconds)
        CONSTRAINT positive_max_suspended_time CHECK (offerings.max_suspended_time >= 0),

    max_inactive_time_sec bigint -- maximum inactive time before channel will be closed
        CONSTRAINT positive_max_inactive_time_sec CHECK (offerings.max_inactive_time_sec > 0),

    free_units smallint NOT NULL DEFAULT 0 -- free units (test, bonus)
        CONSTRAINT positive_free_units CHECK (offerings.free_units >= 0),
    nonce uuid NOT NULL, -- random number to get different hash, with same parameters
    additional_params json -- all additional parameters stored as JSON -- todo: [suggestion] use jsonb to query for parameters
);

-- State channels.
DROP TABLE IF EXISTS channels CASCADE;
CREATE TABLE channels (
    id uuid PRIMARY KEY,
    is_local boolean NOT NULL, -- created locally (by this Client) or retreived (by this Agent)
    agent eth_addr NOT NULL,
    client eth_addr NOT NULL,
    offering uuid NOT NULL REFERENCES offerings(id),
    block int NOT NULL -- block number, when state channel created
        CONSTRAINT positive_block CHECK (channels.block > 0),

    channel_status chan_status NOT NULL, -- status related to blockchain
    service_status svc_status NOT NULL, -- operational status of service
    service_changed_time timestamp with time zone, -- timestamp, when service status changed. Used in aging scenarios. Specifically in suspend -> terminating scenario.
    total_deposit bigint NOT NULL -- total deposit after all top-ups
        CONSTRAINT positive_total_deposit CHECK (channels.total_deposit >= 0),
    salt bigint NOT NULL, -- password salt
    username varchar(100), -- optional username, that can identify service instead of state channel id
    password sha3_256 NOT NULL,
    -- TODO change to bigint
    receipt_balance bigint NOT NULL -- last payment amount received
        CONSTRAINT positive_receipt_balance CHECK (channels.receipt_balance >= 0),
    receipt_signature text NOT NULL -- signature corresponding to last payment
);

-- Client sessions.
DROP TABLE IF EXISTS sessions CASCADE;
CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    channel uuid NOT NULL REFERENCES channels(id),
    started timestamp with time zone NOT NULL, -- time, when session started
    stopped timestamp with time zone, -- time, when session stopped
    units_used bigint NOT NULL -- total units used in this session.
        CONSTRAINT positive_units_used CHECK (sessions.units_used >= 0),

    seconds_consumed bigint NOT NULL -- total seconds interval from started is recorded
        CONSTRAINT positive_seconds_consumed CHECK (sessions.seconds_consumed >= 0),

    last_usage_time timestamp with time zone NOT NULL, -- time of last usage reported
    server_ip inet,
    server_port int
        CONSTRAINT server_port_ct CHECK (sessions.server_port > 0 AND sessions.server_port <= 65535),

    client_ip inet,
    client_port int
        CONSTRAINT client_port_ct CHECK (sessions.client_port > 0 AND sessions.client_port <= 65535)
);

-- Smart contracts.
DROP TABLE IF EXISTS contracts CASCADE;
CREATE TABLE contracts (
    id uuid PRIMARY KEY,
    address sha3_256 NOT NULL, -- ethereum address of contract
    type contract_type NOT NULL,
    version smallint, --version of contract. Greater means newer
    enabled boolean NOT NULL -- contract is in use
);

-- Endpoint messages. Messages that include info about service access.
DROP TABLE IF EXISTS endpoints CASCADE;
CREATE TABLE endpoints (
    id uuid PRIMARY KEY,
    template uuid NOT NULL REFERENCES templates(id), -- corresponding endpoint template
    channel uuid NOT NULL REFERENCES channels(id), -- channel id that is being accessed
    hash sha3_256 NOT NULL, -- message hash
    status msg_status NOT NULL, -- message status
    signature text NOT NULL, -- agent's signature
    payment_receiver_address varchar(106), -- address ("hostname:port") of payment receiver. Can be dns or IP.
    dns varchar(100),
    ip_addr inet,
    username varchar(100),
    password varchar(48),
    additional_params json -- all additional parameters stored as JSON
);

-- Job queue.
DROP TABLE IF EXISTS jobs CASCADE;
CREATE TABLE jobs (
    id uuid PRIMARY KEY,
    type varchar(64) NOT NULL, -- type of task
    status job_status NOT NULL, -- job status
    related_type related_type NOT NULL, -- name of object that relid point on (offering, channel, endpoint, etc.)
    related_id uuid NOT NULL, -- related object (offering, channel, endpoint, etc.)
    created_at timestamp with time zone NOT NULL, -- timestamp, when job was created
    not_before timestamp with time zone NOT NULL, -- timestamp, used to create deffered job
    created_by job_creator NOT NULL, -- job creator
    try_count smallint NOT NULL -- number of tries performed
);

-- Ethereum transactions.
DROP TABLE IF EXISTS eth_txs CASCADE;
CREATE TABLE eth_txs (
    id uuid PRIMARY KEY,
    hash sha3_256 NOT NULL, -- transaction hash
    method text NOT NULL, -- contract method
    status tx_status NOT NULL, -- tx status (custom)
    job uuid REFERENCES jobs(id), -- corresponding job id
    issued timestamp with time zone NOT NULL, -- timestamp, when tx was sent
    block_number_issued bigint
        CONSTRAINT positive_block_number_issued CHECK (eth_txs.block_number_issued > 0), -- block number, when tx was sent to the network

    addr_from eth_addr NOT NULL, -- from ethereum address
    addr_to eth_addr NOT NULL, -- to ethereum address
    nonce numeric, -- tx nonce field
    gas_price bigint
        CONSTRAINT positive_gas_price CHECK (eth_txs.gas_price > 0), -- tx gas_price field

    gas bigint
        CONSTRAINT positive_gas CHECK (eth_txs.gas > 0), -- tx gas field

    tx_raw jsonb -- raw tx as was sent
);

-- Ethereum events.
DROP TABLE IF EXISTS eth_logs CASCADE;
CREATE TABLE eth_logs (
    id uuid PRIMARY KEY,
    tx_hash sha3_256, -- transaction hash
    status tx_status NOT NULL, -- tx status (custom)
    job uuid REFERENCES jobs(id), -- corresponding job id
    block_number bigint
        CONSTRAINT positive_block_number CHECK (eth_logs.block_number > 0),

    addr eth_addr NOT NULL, -- address of contract from which this log originated
    data text NOT NULL, -- contains one or more 32 Bytes non-indexed arguments of the log
    topics jsonb -- array of 0 to 4 32 Bytes DATA of indexed log arguments.
);

END TRANSACTION;
