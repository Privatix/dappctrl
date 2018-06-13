BEGIN TRANSACTION;

-- Test data for integration testing of dappvpn.

INSERT INTO accounts (id, eth_addr, public_key, private_key, name, ptc_balance,
    psc_balance, eth_balance)
VALUES ('e8b17880-8ee5-4fc1-afb2-e6900655d8d5', '', '', '', 'Test channel',
    0, 0, '');

INSERT INTO offerings (id, is_local, tpl, product, hash, status, offer_status,
    block_number_updated, agent, raw_msg, service_name, country, supply,
    unit_name, unit_type, billing_type, setup_price, unit_price, min_units,
    max_unit, billing_interval, max_billing_unit_lag, max_suspended_time,
    free_units)
VALUES ('32000ae1-f752-4d55-8d58-22d05ef08803', true,
    'efc61769-96c8-4c0d-b50a-e4d11fc30523',
    '4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532', '', 'msg_channel_published',
    'register', 1, '', '', 'VPN', 'US', 1, 'megabyte', 'units', 'prepaid', 1, 1,
    1, 100, 1, 0, 0, 0);

INSERT INTO channels (id, agent, client, offering, block, channel_status,
    service_status, total_deposit, salt, password, receipt_balance,
    receipt_signature)
VALUES ('ae5deac9-44c3-4840-bdff-ca9de58c89f4', '', '',
    '32000ae1-f752-4d55-8d58-22d05ef08803', 1, 'active', 'active', 1,
    6012867121110302348, 'JDJhJDEwJHNVbWNtTkVwQk5DMkwuOC5OL1BXU08uYkJMMkxjcmthTW1BZklOTUNjNWZDdWNUOU54Tzlp', 1, 1);

END TRANSACTION;
