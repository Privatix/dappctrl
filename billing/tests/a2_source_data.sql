BEGIN transaction;
-- Truncating all tables from the data that might be present after previous test executions,
-- or even after after other tests.
TRUNCATE TABLE templates CASCADE;
TRUNCATE TABLE offerings CASCADE;
TRUNCATE TABLE products CASCADE;
TRUNCATE TABLE sessions CASCADE;
INSERT INTO templates (id, hash, raw, kind)
VALUES
    (
        '00000000-0000-0000-0000-000000000000',
        'pdgbvdqzyjjtpvprtpetypkhgbarytbxwrmxlwjsfvkb',
        '{}', 'offer'
    );
INSERT INTO products (
    id, name, offer_tpl_id, offer_access_id,
    usage_rep_type, is_server
)
VALUES
    (
        '00000000-0000-0000-0000-000000000000',
        'product', '00000000-0000-0000-0000-000000000000',
        '00000000-0000-0000-0000-000000000000',
        'total', TRUE
    );
INSERT INTO offerings (
    id, is_local, tpl, product, hash, status,
    offer_status, block_number_updated,
    agent, signature, service_name, description,
    country, supply, unit_name, unit_type,
    billing_type, setup_price, unit_price,
    min_units, max_unit, billing_interval,
    max_billing_unit_lag, max_suspended_time,
    max_inactive_time_sec, free_units,
    nonce, additional_params
)
VALUES
    (
        '00000000-0000-0000-0000-000000000000',
        FALSE, '00000000-0000-0000-0000-000000000000',
        '00000000-0000-0000-0000-000000000000',
        'pdgbvdqzyjjtpvprtpetypkhgbarytbxwrmxlwjsfvkb',
        'msg_channel_published', 'register',
        100, '0000000000000000000000000000',
        '0000000000000000000000000001',
        'test service', 'test description',
        'UA', 10, '1', 'seconds', 'postpaid',
        100, 1, 10, 900, 10, 10, 10, 10, 0, '00000000-0000-0000-0000-000000000000',
        '{}'
    );
TRUNCATE TABLE channels CASCADE;
INSERT INTO channels (
    id, is_local, agent, client, offering,
    block, channel_status, service_status,
    service_changed_time, total_deposit,
    salt, username, password, receipt_balance,
    receipt_signature
)
VALUES
    (
        '00000000-0000-0000-0000-000000000001',
        TRUE, '0000000000000000000000000000',
        '0000000000000000000000000001',
        '00000000-0000-0000-0000-000000000000',
        100, 'active', 'active', now(), 10000,
        10, 'test username', 'test password',
        0, 'test signature'
    ),
    (
        '00000000-0000-0000-0000-000000000002',
        TRUE, '0000000000000000000000000000',
        '0000000000000000000000000001',
        '00000000-0000-0000-0000-000000000000',
        100, 'active', 'active', now(), 10000,
        10, 'test username', 'test password',
        0, 'test signature'
    );
INSERT INTO sessions (
    id, channel, started, stopped, units_used,
    seconds_consumed, last_usage_time,
    server_ip, server_port, client_ip,
    client_port
)
VALUES
    (
        '00000000-0000-0000-0000-000000000001',
        '00000000-0000-0000-0000-000000000001',
        now(), now(), 300, 300, now(), '0.0.0.0',
        '3000', '0.0.0.0', '3000'
    ),
    (
        '00000000-0000-0000-0000-000000000002',
        '00000000-0000-0000-0000-000000000001',
        now(), now(), 300, 300, now(), '0.0.0.0',
        '3000', '0.0.0.0', '3000'
    ),
    (
        '00000000-0000-0000-0000-000000000003',
        '00000000-0000-0000-0000-000000000001',
        now(), now(), 300, 300, now(), '0.0.0.0',
        '3000', '0.0.0.0', '3000'
    ),
    (
        '00000000-0000-0000-0000-000000000004',
        '00000000-0000-0000-0000-000000000002',
        now(), now(), 300, 300, now(), '0.0.0.0',
        '3000', '0.0.0.0', '3000'
    ),
    (
        '00000000-0000-0000-0000-000000000005',
        '00000000-0000-0000-0000-000000000002',
        now(), now(), 300, 300, now(), '0.0.0.0',
        '3000', '0.0.0.0', '3000'
    );
END transaction;
