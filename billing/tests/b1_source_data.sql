BEGIN TRANSACTION;

-- Truncating all tables from the data that might be present after previous test executions,
-- or even after after other tests.
TRUNCATE TABLE templates cascade;
TRUNCATE TABLE offerings cascade;
TRUNCATE TABLE products cascade;

INSERT INTO templates
(
  id,
  hash,
  raw,
  kind
)
VALUES
  (
    '00000000-0000-0000-0000-000000000000',
    'pdgbvdqzyjjtpvprtpetypkhgbarytbxwrmxlwjsfvkb',
    '{}',
    'offer'
  );

INSERT INTO products
(
  id,
  name,
  offer_tpl_id,
  offer_access_id,
  usage_rep_type,
  is_server,
  salt,
  password,
  client_ident
)
VALUES
  (
    '00000000-0000-0000-0000-000000000000',
    'product',
    '00000000-0000-0000-0000-000000000000',
    '00000000-0000-0000-0000-000000000000',
    'total',
    TRUE,
    0,
    '7U9gC4AZsSZ9E8NabVkw8nHRlFCJe0o_Yh9qMlIaGAg=',
    'by_channel_id'
  );

INSERT INTO offerings
(
  id,
  is_local,
  tpl,
  product,
  hash,
  status,
  offer_status,
  block_number_updated,
  agent,
  signature,
  service_name,
  description,
  country,
  supply,
  unit_name,
  unit_type,
  billing_type,
  setup_price,
  unit_price,
  min_units,
  max_unit,
  billing_interval,
  max_billing_unit_lag,
  max_suspended_time,
  max_inactive_time_sec,
  free_units,
  additional_params
)
VALUES
  (
    '00000000-0000-0000-0000-000000000000',
    FALSE,
    '00000000-0000-0000-0000-000000000000',
    '00000000-0000-0000-0000-000000000000',
    'pdgbvdqzyjjtpvprtpetypkhgbarytbxwrmxlwjsfvkb',
    'msg_channel_published',
    'register',
    100,
    '0000000000000000000000000000',
    '0000000000000000000000000001',
    'test service',
    'test description',
    'UA',
    10,
    '1',
    'units',
    'postpaid',
    100,
    10,
    10,
    100,
    10,
    10,
    10,
    10,
    0,
    '{}'
  );

TRUNCATE TABLE channels cascade;

INSERT INTO channels
(
  id,
  is_local,
  agent,
  client,
  offering,
  BLOCK,
  channel_status,
  service_status,
  service_changed_time,
  total_deposit,
  salt,
  username,
  password,
  receipt_balance,
  receipt_signature
)
VALUES
  (
    '00000000-0000-0000-0000-000000000001',
    TRUE,
    '0000000000000000000000000000',
    '0000000000000000000000000001',
    '00000000-0000-0000-0000-000000000000',
    100,
    'active',
    'active',
    Now(),
    1, -- NOTE: total_deposit is less than offering setup price.
    10,
    'test username',
    'test password',
    0,
    'test signature'
  ),
  (
    '00000000-0000-0000-0000-000000000002',
    TRUE,
    '0000000000000000000000000000',
    '0000000000000000000000000001',
    '00000000-0000-0000-0000-000000000000',
    100,
    'active',
    'active',
    Now(),
    1000,
    10,
    'test username',
    'test password',
    0,
    'test signature'
  );

END TRANSACTION;
