-- Create new enum type, because altertype doesn't work in transaction block.
CREATE TYPE related_type_v2 AS ENUM (
    'offering', -- service offering
    'channel', -- state channel
    'endpoint', -- service endpoint
    'account', -- for transfer and approve jobs
    'transaction' -- for transaction resends
);

-- Alter type of fields.
ALTER TABLE jobs
  ALTER COLUMN related_type
    SET DATA TYPE related_type_v2
    USING related_type::text::related_type_v2;
ALTER TABLE eth_txs
  ALTER COLUMN related_type
    SET DATA TYPE related_type_v2
    USING related_type::text::related_type_v2;

-- Drop old enum type.
DROP TYPE related_type;
