-- Add related type for rating jobs.
INSERT INTO pg_enum (
    enumtypid, 
    enumlabel, 
    enumsortorder
)
SELECT 
    'related_type'::regtype::oid, 
    'rating', 
    (SELECT MAX(enumsortorder) + 1 FROM pg_enum WHERE enumtypid = 'related_type'::regtype);

-- Channel closing types.
CREATE TYPE closing_type AS ENUM ('coop','uncoop');

-- Closings store closing from blockchain independent from who agents and clients are. Used in rating calculation.
CREATE TABLE closings (
    id uuid PRIMARY KEY,
    type closing_type NOT NULL,
    agent eth_addr NOT NULL, -- eth_addr defined in 00001 up script.
    client eth_addr NOT NULL, -- eth_addr defined in 00001 up script.
    balance bigint, -- close balance in Prix.
    block int NOT NULL -- block number close recorded.
        CONSTRAINT positive_block CHECK (closings.block >= 0)
);

-- Ratings store rating values for an ethereum account.
CREATE TABLE ratings (
    eth_addr eth_addr PRIMARY KEY, -- client or agent ethereum account address.
    val bigint NOT NULL -- accounts rating.
);