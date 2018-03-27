-- Service Units usage reporting type. Can be incremental or total. Indicates how reporting server will report usage of units.
CREATE TYPE usage_rep_type AS ENUM ('incremental', 'total');

-- Templates types.
CREATE TYPE tpl_type AS ENUM ('offer', 'auth', 'access');

-- Billing types.
CREATE TYPE bill_type AS ENUM ('prepaid','postpaid');

-- Unit types. Used for billing calculation.
CREATE TYPE unit_type AS ENUM ('units','seconds');

-- SHA3-256 in base64 (RFC-4648).
CREATE DOMAIN sha3_256 AS char(44);

-- Etehereum address
CREATE DOMAIN eth_addr AS char(20);

 -- Ethereum's uint192 in base64 (RFC-4648).
CREATE DOMAIN privatix_tokens AS char(32);

-- Contract types.
CREATE TYPE contract_type AS ENUM ('ptc','psc');

-- User roles.
CREATE TYPE user_role AS ENUM ('client','agent');

-- Service operational status.
CREATE TYPE svc_status AS ENUM (
    'pending', -- Service is still not fully setup and cannot be used. E.g. waiting for authentication message/endpoint message.
    'active' -- service is now active and can be used.
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
    id uuid PRIMARY KEY,
    public_key text NOT NULL,
    private_key text,
    role user_role NOT NULL, -- agent or client
    is_default boolean, -- default account
    not_inuse boolean -- this account is not in use
);

-- Templates.
CREATE TABLE templates (
    id uuid PRIMARY KEY,
    hash sha3_256 NOT NULL,
    raw json,
    tpl_type tpl_type NOT NULL
);

-- Products. Used to store billing and action related settings.
CREATE TABLE products (
    id uuid PRIMARY KEY,
    name varchar(64) NOT NULL,
    offer_tpl_id uuid REFERENCES templates(id), -- enables product specific billing and actions support for Client
    -- offer_auth_id uuid REFERENCES templates(id), -- currently not in use. for future use.
    offer_access_id uuid REFERENCES templates(id), -- allows to identify endpoint message relation
    usage_rep_type usage_rep_type -- for billing logic. Reporter provides increment or total usage
);

-- Service offerings.
CREATE TABLE offerings (
    id uuid PRIMARY KEY,
    tpl uuid REFERENCES templates(id), -- corresponding template
    product uuid NOT NULL REFERENCES products(id), -- enables product specific billing and actions support for Agent
    hash sha3_256 NOT NULL, -- offering hash
    status msg_status NOT NULL, -- message status
    agent uuid NOT NULL REFERENCES users(id),
    signature text NOT NULL, -- agent's signature
    tpl_version int NOT NULL, -- template version
    service_name varchar(64) NOT NULL, -- name of service
    description text, -- description for UI
    country char(2) NOT NULL, -- ISO 3166-1 alpha-2
    supply int NOT NULL, -- maximum identical offerings for concurrent use through different state channels
    unit_name varchar(10) NOT NULL, -- like megabytes, minutes, etc
    unit_type unit_type NOT NULL, -- type of unit. Time or material.
    billing_type bill_type NOT NULL, -- prepaid/postpaid
    setup_price privatix_tokens, -- setup fee
    unit_price privatix_tokens NOT NULL,
    min_units bigint NOT NULL, -- used to calculate min required deposit
    max_unit bigint, -- optional. If specified automatic termination can be invoked
    billing_interval int NOT NULL, -- every unit numbers, that should be paid, after free units consumed
    max_billing_unit_lag int NOT NULL, --maximum tolerance for payment lag (in units)
    max_suspended_time int NOT NULL, -- maximum time in suspend state, after which service will be terminated (in seconds)
    max_inactive_time_sec bigint, -- maximum inactive time before channel will be closed
    free_units smallint, -- free units (test, bonus)
    nonce uuid NOT NULL, -- random number to get different hash, with same parameters
    additional_params json -- all additional parameters stored as JSON
);

-- State channels.
CREATE TABLE channels (
    id uuid PRIMARY KEY,
    agent uuid NOT NULL REFERENCES users(id),
    client uuid NOT NULL REFERENCES users(id),
    offering uuid NOT NULL REFERENCES offerings(id),
    block int NOT NULL, -- block number, when state channel created
    channel_status chan_status NOT NULL, -- status related to blockchain
    service_status svc_status NOT NULL, -- operational status of service
    service_changed_time timestamp with time zone, -- timestamp, when service status changed. Used in aging scenarios. Specifically in suspend -> terminating scenario.
    total_deposit privatix_tokens NOT NULL, -- total deposit after all top-ups
    salt bigint NOT NULL, -- password salt
    username varchar(100), -- optional username, that can identify service instead of state channel id
    password sha3_256 NOT NULL,
    receipt_balance privatix_tokens NOT NULL, -- last payment amount received
    receipt_signature text NOT NULL -- signature corresponding to last payment
);

-- Client sessions.
CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    channel uuid NOT NULL REFERENCES channels(id),
    started timestamp with time zone NOT NULL, -- time, when session started
    stopped timestamp with time zone, -- time, when session stopped
    units_used bigint NOT NULL, -- total units used in this session.
    seconds_consumed bigint NOT NULL, -- total seconds interval from started is recorded
    last_used_time timestamp with time zone NOT NULL, -- time of last usage reported
    server_ip inet,
    server_port int,
    client_ip inet,
    client_port int
);

-- Smart contracts.
CREATE TABLE contracts (
    id uuid PRIMARY KEY,
    address sha3_256 NOT NULL, -- ethereum address of contract
    type contract_type NOT NULL,
    version smallint, --version of contract. Greater means newer
    enabled boolean NOT NULL -- contract is in use
);

-- Endpoint messages. Messages that include info about service access.
CREATE TABLE endpoints (
    id uuid PRIMARY KEY,
    tpl uuid REFERENCES templates(id), -- corresponding endpoint template
    channel uuid NOT NULL REFERENCES channels(id), -- channel id that is being accessed
    hash sha3_256 NOT NULL, -- message hash
    status msg_status NOT NULL, -- message status
    signature text NOT NULL, -- agent's signature
    tpl_version int NOT NULL, -- template version
    payment_receiver_address varchar(106), -- address ("hostname:port") of payment receiver. Can be dns or IP.
    dns varchar(100),
    ipv4 inet,
    ipv6 inet,
    username varchar(100),
    password varchar(48),
    nonce uuid NOT NULL, -- random number to get different hash, with same parameters
    additional_params json -- all additional parameters stored as JSON
);

-- Job queue.
CREATE TABLE jobs (
    id uuid PRIMARY KEY,
    task_name char(30) NOT NULL, -- name of task
    status job_status NOT NULL, -- job status
    parent_obj char(30) NOT NULL, -- name of object that relid point on (offering, channel, endpoint, etc.)
    relid uuid NOT NULL, -- related object (offering, channel, endpoint, etc.)
    created_at timestamp with time zone NOT NULL, -- timestamp, when job was created
    not_before timestamp with time zone, -- timestamp, used to create deffered job
    created_by job_creator NOT NULL, -- job creator
    fail_count smallint, -- number of failures
    try_count smallint -- number of times job was executed
);

-- Ethereum transactions.
CREATE TABLE ethtxs (
    id uuid PRIMARY KEY,
    hash sha3_256 NOT NULL, -- transaction hash
    method char(30) NOT NULL, -- contract method
    status tx_status NOT NULL, -- tx status (custom)
    job uuid REFERENCES jobs(id) -- corresponding endpoint template
    issued timestamp with time zone NOT NULL, -- timestamp, when tx was sent
    block_number_confirmed int, -- block number, when tx considered confirmed. Always greater or equal to real block number.
    addr_from eth_addr NOT NULL, -- from etehereum address
    addr_to eth_addr NOT NULL, -- from etehereum address
    nonce numeric, -- tx nonce field
    gas_price bigint, -- tx gas_price field
    gas bigint, -- tx gas field
    tx_raw text -- raw tx as was sent
);

-- Ethereum events.
CREATE TABLE ethlog (
    id uuid PRIMARY KEY,
    tx_hash sha3_256, -- transaction hash
    status tx_status NOT NULL -- tx status (custom)
    job uuid REFERENCES jobs(id) -- corresponding endpoint template
    block_number int, -- event block number field
    block_number_confirmed int, -- block number, when tx considered confirmed. Always greater or equal to real block number.
    addr eth_addr NOT NULL, -- address from which this log originated
    data text, -- contains one or more 32 Bytes non-indexed arguments of the log
    topics text -- array of 0 to 4 32 Bytes DATA of indexed log arguments.
);
