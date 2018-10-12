--templateid = efc61769-96c8-4c0d-b50a-e4d11fc30523
BEGIN TRANSACTION;

INSERT INTO templates (id, hash, raw, kind)
VALUES ('efc61769-96c8-4c0d-b50a-e4d11fc30523',
        'HGuVky1SotjobyIVpiGw4jBvFNt28MtF5uNF7OCOYdo=',
        '{
            "schema": {
                "properties": {
                    "additionalParams": {
                        "default": {},
                        "minUploadMbps": {
                            "title": "minimum upload speed (Mbps)",
                            "type": "number"
                        },
                        "minDownloadMbps": {
                            "title": "minimum download speed (Mbps)",
                            "type": "number"
                        },
                        "type": "object"
                    },
                    "agent": {
                        "title": "agent uuid",
                        "type": "string"
                    },
                    "billingInterval": {
                        "title": "billing interval",
                        "type": "number"
                    },
                    "billingType": {
                        " enumNames": [
                            "prepaid",
                            "postpaid"
                        ],
                        "enum": [
                            "prepaid",
                            "postpaid"
                        ],
                        "title": "billing type",
                        "type": "string"
                    },
                    "country": {
                        "title": "country",
                        "type": "string"
                    },
                    "freeUnits": {
                        "title": "free units",
                        "type": "number"
                    },
                    "maxBillingUnitLag": {
                        "title": "max billing unit lag",
                        "type": "number"
                    },
                    "maxSuspendTime": {
                        "title": "max suspend time",
                        "type": "number"
                    },
                    "minUnits": {
                        "title": "min units",
                        "type": "number"
                    },
                    "product": {
                        "default": "1",
                        "type": "string"
                    },
                    "serviceName": {
                        "title": "Name of service (e.g. VPN)",
                        "type": "string"
                    },
                    "setupPrice": {
                        "title": "setup fee",
                        "type": "number"
                    },
                    "supply": {
                        "title": "service supply",
                        "type": "number"
                    },
                    "template": {
                        "default": "1",
                        "type": "string"
                    },
                    "unitName": {
                        "title": "like megabytes, minutes, etc",
                        "type": "string"
                    },
                    "unitPrice": {
                        "title": "unit price",
                        "type": "number"
                    },
                    "unitType": {
                        "title": "service unit",
                        "type": "number"
                    }
                },
                "required": [
                    "serviceName",
                    "supply",
                    "unitName",
                    "unitType",
                    "billingType",
                    "setupPrice",
                    "unitPrice",
                    "country",
                    "minUnits",
                    "billingInterval",
                    "maxBillingUnitLag",
                    "freeUnits",
                    "template",
                    "product",
                    "agent",
                    "additionalParams",
                    "maxSuspendTime"
                ],
                "title": "Privatix VPN offering",
                "type": "object"
            },
            "uiSchema": {
                "additionalParams": {
                    "ui:widget": "hidden"
                },
                "agent": {
                    "ui:widget": "hidden"
                },
                "billingInterval": {
                    "ui:help": "Specified in unit_of_service. Represent, how often Client MUST provide payment approval to Agent."
                },
                "billingType": {
                    "ui:help": "prepaid/postpaid"
                },
                "country": {
                    "ui:help": "Country of service endpoint in ISO 3166-1 alpha-2 format."
                },
                "freeUnits": {
                    "ui:help": "Used to give free trial, by specifying how many intervals can be consumed without payment"
                },
                "maxBillingUnitLag": {
                    "ui:help": "Maximum payment lag in units after, which Agent will suspend serviceusage."
                },
                "maxSuspendTime": {
                    "ui:help": "Maximum time without service usage. Agent will consider, that Client will not use service and stop providing it. Period is specified in minutes."
                },
                "minUnits": {
                    "ui:help": "Used to calculate minimum deposit required"
                },
                "product": {
                    "ui:widget": "hidden"
                },
                "serviceName": {
                    "ui:help": "enter name of service"
                },
                "setupPrice": {
                    "ui:help": "setup fee"
                },
                "supply": {
                    "ui:help": "Maximum supply of services according to service offerings. It represents maximum number of clients that can consume this service offering concurrently."
                },
                "template": {
                    "ui:widget": "hidden"
                },
                "unitName": {
                    "ui:help": "MB/Minutes"
                },
                "unitPrice": {
                    "ui:help": "PRIX that must be paid for unit_of_service"
                },
                "unitType": {
                    "ui:help": "units or seconds"
                }
            }
        }',
        'offer');

INSERT INTO templates (id, hash, raw, kind)
VALUES ('d0dfbbb2-dd07-423a-8ce0-1e74ce50105b',
        'RJM57hqcmEdDcxi-rahi5m5lKs6ISo5Oa0l67cQwmTQ=',
        '{
            "definitions": {
                "host": {
                "pattern": "^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])(\\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9]))*:[0-9]{2,5}$",
                "type": "string"
                },
                "simple_url": {
		        "pattern": "^(http:\\/\\/www\\.|https:\\/\\/www\\.|http:\\/\\/|https:\\/\\/)?.+",
                "type": "string"
                },
                "uuid": {
                "pattern": "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
                "type": "string"
                }
            },
            "properties": {
                "additionalParams": {
                    "additionalProperties": {
                        "type": "string"
                    },
                    "minProperties": 1,
                    "type": "object"
                },
                "password": {
                    "type": "string"
                },
                "paymentReceiverAddress": {
                    "$ref": "#/definitions/simple_url"
                },
                "serviceEndpointAddress": {
                    "type": "string"
                },
                "templateHash": {
                    "type": "string"
                },
                "username": {
                    "$ref": "#/definitions/uuid"
                }
            },
            "required": [
                "templateHash",
                "paymentReceiverAddress",
                "serviceEndpointAddress",
                "additionalParams"
            ],
            "title": "Privatix VPN access",
            "type": "object"
	    }',
        'access');

INSERT INTO products (id, name, offer_tpl_id, offer_access_id, usage_rep_type,
                      is_server, salt, password, client_ident, config, service_endpoint_address)
VALUES ('4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532', 'VPN server',
        'efc61769-96c8-4c0d-b50a-e4d11fc30523', 'd0dfbbb2-dd07-423a-8ce0-1e74ce50105b',
        'total', TRUE, 6012867121110302348,
        'JDJhJDEwJHNVbWNtTkVwQk5DMkwuOC5OL1BXU08uYkJMMkxjcmthTW1BZklOTUNjNWZDdWNUOU54Tzlp',
        'by_channel_id', '{"somekey": "somevalue"}', '127.0.0.1');

INSERT INTO products(	id, name, offer_tpl_id, offer_access_id, usage_rep_type,
                      is_server, salt, password, client_ident, config, service_endpoint_address)
	VALUES ('37fa14e4-51da-4021-8ea9-9e725229e8aa', 'VPN client',
     'efc61769-96c8-4c0d-b50a-e4d11fc30523', 'd0dfbbb2-dd07-423a-8ce0-1e74ce50105b',
     'total', false, 6012867121110302348,
     'JDJhJDEwJHNVbWNtTkVwQk5DMkwuOC5OL1BXU08uYkJMMkxjcmthTW1BZklOTUNjNWZDdWNUOU54Tzlp',
     'by_channel_id', '{"somekey": "somevalue"}', '127.0.0.1');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.min.confirmations',
        '1',
        2,
        'have value (stored as string) that is null or integer and' ||
        ' represents how many ethereum blocks should be mined after' ||
        ' block where transaction of interest exists. As there is non' ||
        ' zero probability of attack where some last blocks can be' ||
        ' generated by attacker and will be than ignored by ethereum' ||
        ' network (uncle blocks) after attack detection. dappctrl' ||
        ' give ability to user to specify how many latest blocks' ||
        ' are considered non reliable. These last blocks' ||
        ' will not be used to fetch events or transactions.',
        'ethereum confirmation blocks');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.freshblocks',
        '11520',
        2,
        'defines number of latest ethereum blocks to retrieve.' ||
        ' If eth.event.freshblocks is null or zero then all events' ||
        ' will be downloaded.',
        'last events blocks');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.blocklimit',
        '80',
        2,
        'maximum number of ethereum blocks that is used to scan' ||
        ' for new events. It is used as pagination mechanism while' ||
        ' querying ethereum JSON RPC. If eth.event.blocklimit is null' ||
        ' or zero then no pagination is used, which is not recommended.',
        'maximum events blocks');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('error.sendremote',
        'true',
        2,
        'Allow error reporting to send logs to Privatix.',
        'error reporting');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.default.gasprice',
        '20000000000',
        2,
        'Default GAS price for transactions.',
        'default gas price');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.max.deposit',
        '30000000000',
        2,
        'We temporarily limit total token deposits in a channel to 300' ||
        ' PRIX. This is just for the bug bounty release, as a safety measure.',
        'maximum deposit');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.lastProcessedBlock',
        '0',
        2,
        'Last block number in blockchain stores last proccessed block.',
        'last processed block');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('system.version.db',
        '0.12.0',
        1,
        'Version of database.',
        'db version');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('offering.autopopup',
        'true',
        2,
        'Allow offerings to pop up automatically.',
        'offering autopopup');

END TRANSACTION;
