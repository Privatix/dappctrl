-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE log_events;
DROP TYPE log_level;
DROP TABLE eth_logs;
DROP TABLE eth_txs;
DROP TABLE jobs;
DROP TABLE endpoints;
DROP TABLE contracts;
DROP TABLE sessions;
DROP TABLE channels;
DROP TABLE offerings;
DROP TABLE products;
DROP TABLE templates;
DROP TABLE users;
DROP TABLE accounts;
DROP TABLE settings;
DROP TYPE related_type;
DROP TYPE job_status;
DROP TYPE job_creator;
DROP TYPE tx_status;
DROP TYPE offer_status;
DROP TYPE msg_status;
DROP TYPE chan_status;
DROP TYPE svc_status;
DROP DOMAIN eth_addr;
DROP DOMAIN bcrypt_hash;
DROP DOMAIN tx_hash_hex;
DROP DOMAIN sha3_256;
DROP TYPE client_ident_type;
DROP TYPE contract_type;
DROP TYPE unit_type;
DROP TYPE bill_type;
DROP TYPE tpl_kind;
DROP TYPE usage_rep_type;

