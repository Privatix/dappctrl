BEGIN TRANSACTION;

-- Test data for integration testing of dappvpn.

INSERT INTO products (id, name, usage_rep_type, is_server, salt, password,
    client_ident)
VALUES ('4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532', 'Test VPN service', 'total',
    true, 6012867121110302348, '7U9gC4AZsSZ9E8NabVkw8nHRlFCJe0o_Yh9qMlIaGAg=',
    'by_channel_id');

INSERT INTO accounts (id, eth_addr, public_key, private_key, name, ptc_balance,
    psc_balance, eth_balance)
VALUES ('e8b17880-8ee5-4fc1-afb2-e6900655d8d5', '', '', '', 'Test channel',
    0, 0, '');

INSERT INTO templates (id, hash, raw, kind)
VALUES ('58fbe052-3f34-4b17-88c0-1121a8cf9336', '', '{}', 'offer');

INSERT INTO offerings (id, is_local, tpl, product, hash, status, offer_status,
    block_number_updated, agent, signature, service_name, country, supply,
    unit_name, unit_type, billing_type, setup_price, unit_price, min_units,
    max_unit, billing_interval, max_billing_unit_lag, max_suspended_time,
    free_units)
VALUES ('32000ae1-f752-4d55-8d58-22d05ef08803', true,
    '58fbe052-3f34-4b17-88c0-1121a8cf9336',
    '4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532', '', 'msg_channel_published',
    'register', 1, '', '', 'VPN', 'US', 1, 'megabyte', 'units', 'prepaid', 1,
    1, 1, 100, 1, 0, 0, 0);

INSERT INTO channels (id, is_local, agent, client, offering, block,
    channel_status, service_status, total_deposit, salt, password,
    receipt_balance, receipt_signature)
VALUES ('ae5deac9-44c3-4840-bdff-ca9de58c89f4', true, '', '',
    '32000ae1-f752-4d55-8d58-22d05ef08803', 1, 'active', 'active', 1,
    6012867121110302348, '7U9gC4AZsSZ9E8NabVkw8nHRlFCJe0o_Yh9qMlIaGAg=', 1, 1);

-- Test data for Endpoint Message template.

INSERT INTO templates (id, hash, raw, kind)
VALUES ('d2ea83d3-8513-45e4-90ae-d9ab20406f32', '',
    '{
        "title": "Endpoint Message template",
        "type": "object",
        "definitions": {
            "uuid": {
                "type": "string",
                "pattern":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
            },
            "host": {
                "type": "string",
                "pattern":"^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])(\\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9]))*:[0-9]{2,5}$"
            }
        },
        "properties": {
            "templateHash": {
                "$ref":"#/definitions/uuid"
            },
            "username":  {
                "$ref":"#/definitions/uuid"
            },
            "password": {
                "type": "string"
            },
            "paymentReceiverAddress": {
                "$ref":"#/definitions/host"
            },
            "serviceEndpointAddress": {
                "type": "string"
            },
            "additionalParams": {
                "type": "object",
                "minProperties": 1,
                "additionalProperties": {
                    "type": "string"
                }
            }
        },
        "required": ["templateHash", "paymentReceiverAddress", "serviceEndpointAddress", "additionalParams"]
    }', 'offer');

INSERT INTO products (id, name, offer_access_id, usage_rep_type, is_server,
    salt, password, client_ident, config)
VALUES ('8d6eec23-0e61-41a7-a6e2-2c16fb2caddb', 'Test Endpoint Message',
    'd2ea83d3-8513-45e4-90ae-d9ab20406f32', 'total', true, 6012867121110302348,
    '7U9gC4AZsSZ9E8NabVkw8nHRlFCJe0o_Yh9qMlIaGAg=', 'by_channel_id',
    '{"ca":"crt.valid","caData":"-----BEGIN CERTIFICATE-----\nMIIDnzCCAyWgAwIBAgIQWyXOaQfEJlVm0zkMmalUrTAKBggqhkjOPQQDAzCBhTEL\nMAkGA1UEBhMCR0IxGzAZBgNVBAgTEkdyZWF0ZXIgTWFuY2hlc3RlcjEQMA4GA1UE\nBxMHU2FsZm9yZDEaMBgGA1UEChMRQ09NT0RPIENBIExpbWl0ZWQxKzApBgNVBAMT\nIkNPTU9ETyBFQ0MgQ2VydGlmaWNhdGlvbiBBdXRob3JpdHkwHhcNMTQwOTI1MDAw\nMDAwWhcNMjkwOTI0MjM1OTU5WjCBkjELMAkGA1UEBhMCR0IxGzAZBgNVBAgTEkdy\nZWF0ZXIgTWFuY2hlc3RlcjEQMA4GA1UEBxMHU2FsZm9yZDEaMBgGA1UEChMRQ09N\nT0RPIENBIExpbWl0ZWQxODA2BgNVBAMTL0NPTU9ETyBFQ0MgRG9tYWluIFZhbGlk\nYXRpb24gU2VjdXJlIFNlcnZlciBDQSAyMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcD\nQgAEAjgZgTrJaYRwWQKOqIofMN+83gP8eR06JSxrQSEYgur5PkrkM8wSzypD/A7y\nZADA4SVQgiTNtkk4DyVHkUikraOCAWYwggFiMB8GA1UdIwQYMBaAFHVxpxlIGbyd\nnepBR9+UxEh3mdN5MB0GA1UdDgQWBBRACWFn8LyDcU/eEggsb9TUK3Y9ljAOBgNV\nHQ8BAf8EBAMCAYYwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHSUEFjAUBggrBgEF\nBQcDAQYIKwYBBQUHAwIwGwYDVR0gBBQwEjAGBgRVHSAAMAgGBmeBDAECATBMBgNV\nHR8ERTBDMEGgP6A9hjtodHRwOi8vY3JsLmNvbW9kb2NhLmNvbS9DT01PRE9FQ0ND\nZXJ0aWZpY2F0aW9uQXV0aG9yaXR5LmNybDByBggrBgEFBQcBAQRmMGQwOwYIKwYB\nBQUHMAKGL2h0dHA6Ly9jcnQuY29tb2RvY2EuY29tL0NPTU9ET0VDQ0FkZFRydXN0\nQ0EuY3J0MCUGCCsGAQUFBzABhhlodHRwOi8vb2NzcC5jb21vZG9jYTQuY29tMAoG\nCCqGSM49BAMDA2gAMGUCMQCsaEclgBNPE1bAojcJl1pQxOfttGHLKIoKETKm4nHf\nEQGJbwd6IGZrGNC5LkP3Um8CMBKFfI4TZpIEuppFCZRKMGHRSdxv6+ctyYnPHmp8\n7IXOMCVZuoFwNLg0f+cB0eLLUg==\n-----END CERTIFICATE-----\n","caPathName":"samples/crt.valid","cipher":"AES-256-CBC","comp-lzo":"","connect-retry":"2 120","keepalive":"10 120","ping":"10","ping-restart":"10","proto":"udp"}');


INSERT INTO offerings (id, is_local, tpl, product, hash, status, offer_status,
    block_number_updated, agent, signature, service_name, country, supply,
    unit_name, unit_type, billing_type, setup_price, unit_price, min_units,
    max_unit, billing_interval, max_billing_unit_lag, max_suspended_time,
    free_units)
VALUES ('37417f97-0d3f-416c-b8e5-91fdd40086cb', false,
    'd2ea83d3-8513-45e4-90ae-d9ab20406f32',
    '8d6eec23-0e61-41a7-a6e2-2c16fb2caddb', '', 'msg_channel_published',
    'register', 1, '', '', 'VPN', 'US', 1, 'megabyte', 'units', 'prepaid', 1,
    1, 1, 100, 1, 0, 0, 0);

INSERT INTO channels (id, is_local, agent, client, offering, block,
    channel_status, service_status, total_deposit, salt, password,
    receipt_balance, receipt_signature)
VALUES ('902871e1-bb92-48d2-8932-350f88ce1e61', false, '', '',
    '37417f97-0d3f-416c-b8e5-91fdd40086cb', 1, 'active', 'active', 1,
    6012867121110302348, '7U9gC4AZsSZ9E8NabVkw8nHRlFCJe0o_Yh9qMlIaGAg=', 1, 1);

END TRANSACTION;
