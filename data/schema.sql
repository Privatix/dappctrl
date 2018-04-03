test
-- Subjects are party in distributed trade.
-- Each of them can play an agent role, a client role, or both of them.
CREATE TABLE subjects (
    id uuid PRIMARY KEY,
    public_key text NOT NULL,
    private_key text
);

-- Service types.
CREATE TYPE service AS ENUM ('vpn');

-- Service offerings.
CREATE TABLE offerings(
    id uuid PRIMARY KEY,
    agent uuid NOT NULL REFERENCES subjects(id),
    service service NOT NULL,
    supply int NOT NULL,
    signature text NOT NULL
);

-- SHA3-256 in base64 (RFC-4648).
CREATE DOMAIN sha3_256 AS char(44);

 -- Ethereum's uint192 in base64 (RFC-4648).
CREATE DOMAIN privatix_tokens AS char(32);

-- VPN protocols.
CREATE TYPE protocol AS ENUM ('tcp', 'udp');

-- VPN service offerings.
CREATE TABLE offerings_vpn(
    id uuid PRIMARY KEY REFERENCES offerings(id),
    hash sha3_256 NOT NULL,
    country char(2) NOT NULL,
    upload_price privatix_tokens NOT NULL,
    download_price privatix_tokens NOT NULL,
    min_upload bigint NOT NULL,
    max_upload bigint NOT NULL,
    min_download bigint NOT NULL,
    max_download bigint NOT NULL,
    billing_interval int NOT NULL,
    billing_deviation int NOT NULL,
    free_intervals smallint NOT NULL,
    min_upload_bps bigint NOT NULL,
    min_download_bps bigint NOT NULL,
    template_version smallint NOT NULL,
    protocol protocol NOT NULL
);

-- State channel states.
CREATE TYPE channel_state AS ENUM (
    'open',
    'closing',
    'closed_coop',
    'closed_uncoop'
);

-- State channels.
CREATE TABLE channels (
    id uuid PRIMARY KEY,
    agent uuid NOT NULL REFERENCES subjects(id),
    client uuid NOT NULL REFERENCES subjects(id),
    offering uuid NOT NULL REFERENCES offerings(id),
    block int NOT NULL,
    state channel_state NOT NULL,
    total_deposit privatix_tokens NOT NULL,
    closed_deposit privatix_tokens NOT NULL,
    salt bigint NOT NULL,
    password sha3_256 NOT NULL,
    receipt_balance privatix_tokens NOT NULL,
    receipt_signature text NOT NULL
);

-- Client sessions.
CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    channel uuid NOT NULL REFERENCES channels(id),
    started timestamp with time zone NOT NULL,
    stopped timestamp with time zone
);

-- Client sessions for VPN service.
CREATE TABLE sessions_vpn (
    id uuid PRIMARY KEY REFERENCES sessions(id),
    server_ip inet NOT NULL,
    client_ip inet NOT NULL,
    client_port int NOT NULL,
    uploaded bigint NOT NULL,
    downloaded bigint NOT NULL
);
