-- Create new enum type, because altertype doesn't work in transaction block.
CREATE TYPE related_type AS ENUM (
    'offering', -- service offering
    'channel', -- state channel
    'endpoint', -- service endpoint
    'account' -- for transfer and approve jobs
);

-- Alter type of fields.
ALTER TABLE jobs
  ALTER COLUMN related_type
    SET DATA TYPE related_type
    USING related_type::text::related_type;
ALTER TABLE eth_txs
  ALTER COLUMN related_type
    SET DATA TYPE related_type
    USING related_type::text::related_type;

-- Drop old enum type.
DROP TYPE related_type_v2;
