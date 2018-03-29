-- While Postgres out-of-the-box supports storing UUID,
-- generating UUID values requires an extension.
-- It is distributed with the vanilla postgres itself,
-- but is not enabled by default.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Service Units usage reporting type. Can be incremental or total. Indicates how reporting server will report usage of units.
CREATE TYPE usage_rep_type AS ENUM ('incremental', 'total');

-- Templates kinds.
CREATE TYPE tpl_kind AS ENUM ('offer', 'auth', 'access');

-- Billing types.
CREATE TYPE bill_type AS ENUM ('prepaid','postpaid');

-- Unit types. Used for billing calculation.
CREATE TYPE unit_type AS ENUM ('units','seconds');

-- SHA3-256 in base64 (RFC-4648).
CREATE DOMAIN sha3_256 AS char(44);

-- Etehereum address
CREATE DOMAIN eth_addr AS char(28);

 -- Ethereum's uint192 in base64 (RFC-4648).
CREATE DOMAIN privatix_tokens AS char(32);

-- Contract types.
CREATE TYPE contract_type AS ENUM ('ptc','psc');

-- User kinds.
CREATE TYPE user_kind AS ENUM ('client','agent');

-- Service operational status.
CREATE TYPE svc_status AS ENUM (
    'pending', -- Service is still not fully setup and cannot be used. E.g. waiting for authentication message/endpoint message.
    'active', -- service is now active and can be used.
    'suspended', -- service usage is not allowed. Usually used to temporary disallow access.
    'terminated' -- service is permanently deactivated.
);

-- State channel states.
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
CREATE TYPE msg_status AS ENUM (
	'unpublished', -- saved in DB, but not published
	'bchain_publishing', -- publishing in blockchain
	'bchain_published', -- published in blockchain
	'msg_channel_published' -- published in messaging channel
);

-- Transaction statuses.
CREATE TYPE tx_status AS ENUM (
	'unsent', -- saved in DB, but not sent
	'sent', -- sent w/o error to eth node
	'mined', -- tx mined
	'uncle' -- tx is went to uncle block
);

-- Job creator.
CREATE TYPE job_creator AS ENUM (
	'user', -- by user through UI
	'billing_checker', -- by billing checker procedure
	'bc_monitor', -- by blockchain monitor
	'task' -- by another task
);

-- Job status.
CREATE TYPE job_status AS ENUM (
	'new', -- previously never executed
	'failed', -- failed to sucessfully execute
	'skipped', -- skipped by user
	'done' -- successfully executed
);

-- Users are party in distributed trade.
-- Each of them can play an agent role, a client role, or both of them.
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    public_key text NOT NULL,
    private_key text,
    kind user_kind NOT NULL, -- agent or client
    is_default BOOLEAN DEFAULT FALSE, -- default account
    in_use BOOLEAN DEFAULT TRUE-- this account is in use or not
);

-- Templates.
CREATE TABLE templates (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    hash sha3_256 NOT NULL,
    raw json NOT NULL,
    kind tpl_kind NOT NULL
);

-- Products. Used to store billing and action related settings.
CREATE TABLE products (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(64) NOT NULL, -- todo: [suggestion] change type to "text"
    offer_tpl_id uuid REFERENCES templates(id), -- enables product specific billing and actions support for Client
    -- offer_auth_id uuid REFERENCES templates(id), -- currently not in use. for future use.
    offer_access_id uuid REFERENCES templates(id), -- allows to identify endpoint message relation
    usage_rep_type usage_rep_type NOT NULL -- for billing logic. Reporter provides increment or total usage
);

-- Service offerings.
CREATE TABLE offerings (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tpl uuid REFERENCES templates(id), -- corresponding template
    product uuid NOT NULL REFERENCES products(id), -- enables product specific billing and actions support for Agent
    hash sha3_256 NOT NULL, -- offering hash
    status msg_status NOT NULL, -- message status
    agent uuid NOT NULL REFERENCES users(id),
    signature text NOT NULL, -- agent's signature
    tpl_version INT NOT NULL, -- template version
    service_name varchar(64) NOT NULL, -- name of service -- todo: [suggestion] change type to "text"
    description text, -- description for UI
    country char(2) NOT NULL, -- ISO 3166-1 alpha-2
    supply INT NOT NULL, -- maximum identical offerings for concurrent use through different state channels
    unit_name varchar(10) NOT NULL, -- like megabytes, minutes, etc -- todo: [suggestion] change type to "text"
    unit_type unit_type NOT NULL, -- type of unit. Time or material.
    billing_type bill_type NOT NULL, -- prepaid/postpaid
    setup_price privatix_tokens, -- setup fee
    unit_price privatix_tokens NOT NULL,
    min_units BIGINT NOT NULL -- used to calculate min required deposit
        CONSTRAINT positive_min_units CHECK (offerings.min_units >= 0),

    max_unit BIGINT -- optional. If specified automatic termination can be invoked
        CONSTRAINT positive_max_unit CHECK (offerings.max_unit >= 0),

    billing_interval INT NOT NULL -- every unit numbers, that should be paid, after free units consumed
        CONSTRAINT positive_billing_interval CHECK (offerings.billing_interval > 0),

    max_billing_unit_lag INT NOT NULL --maximum tolerance for payment lag (in units)
        CONSTRAINT positive_max_billing_unit_lag CHECK (offerings.max_billing_unit_lag > 0),

    max_suspended_time INT NOT NULL -- maximum time in suspend state, after which service will be terminated (in seconds)
        CONSTRAINT positive_max_suspended_time CHECK (offerings.max_suspended_time >= 0),

    max_inactive_time_sec BIGINT -- maximum inactive time before channel will be closed
        CONSTRAINT positive_max_inactive_time_sec CHECK (offerings.max_inactive_time_sec > 0),

    free_units SMALLINT NOT NULL DEFAULT 0 -- free units (test, bonus)
        CONSTRAINT positive_free_units CHECK (offerings.free_units > 0),

    nonce uuid NOT NULL, -- random number to get different hash, with same parameters -- todo: [suggestion] change type to "bigint"
    additional_params json -- all additional parameters stored as JSON -- todo: [suggestion] change type to "jsonb" (it is significantly faster and supports indexing)
);

-- State channels.
CREATE TABLE channels (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent uuid NOT NULL REFERENCES users(id),
    client uuid NOT NULL REFERENCES users(id),
    offering uuid NOT NULL REFERENCES offerings(id),
    block INT NOT NULL -- block number, when state channel created
        CONSTRAINT positive_block CHECK (channels.block > 0),

    channel_status chan_status NOT NULL, -- status related to blockchain
    service_status svc_status NOT NULL, -- operational status of service
    service_changed_time TIMESTAMP WITH TIME ZONE, -- timestamp, when service status changed. Used in aging scenarios. Specifically in suspend -> terminating scenario.
    total_deposit privatix_tokens NOT NULL, -- total deposit after all top-ups
    salt BIGINT NOT NULL, -- password salt
    username varchar(100), -- optional username, that can identify service instead of state channel id -- todo: [suggestion] change type to "text"
    password sha3_256 NOT NULL, -- todo: [suggestion] rename field to "password_hash"
    receipt_balance privatix_tokens NOT NULL, -- last payment amount received
    receipt_signature text NOT NULL -- signature corresponding to last payment
);

-- Client sessions.
CREATE TABLE sessions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel uuid NOT NULL REFERENCES channels(id),
    started TIMESTAMP WITH TIME ZONE NOT NULL, -- time, when session started
    stopped TIMESTAMP WITH TIME ZONE, -- time, when session stopped
    units_used BIGINT NOT NULL -- total units used in this session.
        CONSTRAINT positive_units_used CHECK (sessions.units_used >= 0),

    seconds_consumed BIGINT NOT NULL -- total seconds interval from started is recorded
        CONSTRAINT positive_seconds_consumed CHECK (sessions.seconds_consumed > 0),

    last_usage_time TIMESTAMP WITH TIME ZONE NOT NULL, -- time of last usage reported
    server_ip inet,
    server_port INT
        CONSTRAINT server_port_ct CHECK (sessions.server_port > 0 AND sessions.server_port <= 65535),

    client_ip inet,
    client_port INT
        CONSTRAINT client_port_ct CHECK (sessions.client_port > 0 AND sessions.client_port <= 65535)
);

-- Smart contracts.
CREATE TABLE contracts (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    address sha3_256 NOT NULL, -- ethereum address of contract
    type contract_type NOT NULL,
    version SMALLINT, --version of contract. Greater means newer -- todo: [suggestion] add constraint
    enabled boolean NOT NULL -- contract is in use -- todo: [suggestion] add default value
);

-- Endpoint messages. Messages that include info about service access.
CREATE TABLE endpoints (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tpl uuid REFERENCES templates(id), -- corresponding endpoint template
    channel uuid NOT NULL REFERENCES channels(id), -- channel id that is being accessed
    hash sha3_256 NOT NULL, -- message hash
    status msg_status NOT NULL, -- message status
    signature text NOT NULL, -- agent's signature
    tpl_version INT NOT NULL, -- template version
    payment_receiver_address varchar(106), -- address ("hostname:port") of payment receiver. Can be dns or IP. -- todo: [suggestion] change type to "text"
    dns varchar(100), -- todo: [suggestion] change type to "text"
    ip_addr inet,
    username varchar(100), -- todo: [suggestion] change type to "text"
    password varchar(48), -- todo: [suggestion] change type to "text"
    additional_params json -- all additional parameters stored as JSON -- todo: [suggestion] change type to jsonb
);

-- Job queue.
CREATE TABLE jobs (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_name text NOT NULL, -- name of task
    status job_status NOT NULL, -- job status
    parent_obj text NOT NULL, -- name of object that relid point on (offering, channel, endpoint, etc.)
    rel_id uuid NOT NULL, -- related object (offering, channel, endpoint, etc.)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL, -- timestamp, when job was created
    not_before TIMESTAMP WITH TIME ZONE, -- timestamp, used to create deffered job
    created_by job_creator NOT NULL, -- job creator
    fail_count SMALLINT DEFAULT 0
        CONSTRAINT positive_fail_count CHECK (jobs.fail_count >= 0), -- number of failures

    attempts_count SMALLINT DEFAULT 0
        CONSTRAINT positive_attempts_count CHECK (jobs.attempts_count >= 0) -- number of times job was executed
);

-- Ethereum transactions.
CREATE TABLE eth_txs (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    hash sha3_256 NOT NULL, -- transaction hash
    method text NOT NULL, -- contract method
    status tx_status NOT NULL, -- tx status (custom)
    job uuid REFERENCES jobs(id), -- corresponding endpoint template
    issued TIMESTAMP WITH TIME ZONE NOT NULL, -- timestamp, when tx was sent
    block_number_issued BIGINT
        CONSTRAINT positive_block_number_issued CHECK (eth_txs.block_number_issued > 0), -- block number, when tx was sent to the network

    addr_from eth_addr NOT NULL, -- from ethereum address
    addr_to eth_addr NOT NULL, -- from ethereum address
    nonce numeric, -- tx nonce field
    gas_price BIGINT
        CONSTRAINT positive_gas_price CHECK (eth_txs.gas_price > 0), -- tx gas_price field

    gas BIGINT
        CONSTRAINT positive_gas CHECK (eth_txs.gas > 0), -- tx gas field

    tx_raw jsonb -- raw tx as was sent
);


-- Ethereum events.
CREATE TABLE eth_logs (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tx_hash sha3_256, -- transaction hash
    status tx_status NOT NULL, -- tx status (custom)
    job uuid REFERENCES jobs(id), -- corresponding endpoint template
    block_number BIGINT
        CONSTRAINT positive_block_number CHECK (eth_logs.block_number > 0),

    addr eth_addr NOT NULL, -- address from which this log originated
    data text NOT NULL, -- contains one or more 32 Bytes non-indexed arguments of the log
    topics jsonb -- array of 0 to 4 32 Bytes DATA of indexed log arguments.
);

-- todo: scheme does not contains any index. We should definitely add them.
