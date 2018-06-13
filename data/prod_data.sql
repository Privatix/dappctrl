BEGIN TRANSACTION;

INSERT INTO templates (id, hash, raw, kind)
VALUES ('efc61769-96c8-4c0d-b50a-e4d11fc30523',
        'HGuVky1SotjobyIVpiGw4jBvFNt28MtF5uNF7OCOYdo=',
        '{
            "schema": {
                "properties": {
                    "additionalParams": {
                        "default": "{}",
                        "type": "string"
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
                "title": "VPN Service Offering",
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
            "title": "Endpoint Message template",
            "type": "object"
	    }',
        'access');

INSERT INTO products (id, name, offer_tpl_id, offer_access_id, usage_rep_type,
                      is_server, salt, password, client_ident, config, service_endpoint_address)
VALUES ('4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532', 'VPN service',
        'efc61769-96c8-4c0d-b50a-e4d11fc30523', 'd0dfbbb2-dd07-423a-8ce0-1e74ce50105b',
        'total', TRUE, 6012867121110302348,
        'JDJhJDEwJHNVbWNtTkVwQk5DMkwuOC5OL1BXU08uYkJMMkxjcmthTW1BZklOTUNjNWZDdWNUOU54Tzlp', 'by_channel_id', '{"somekey": "somevalue"}', 'localhost');

INSERT INTO settings (key, value, description, name)
VALUES ('eth.min.confirmations',
        '1',
        'have value (stored as string) that is null or integer and represents how many ethereum blocks should be mined after block where transaction of interest exists. As there is non zero probability of attack where some last blocks can be generated by attacker and will be than ignored by ethereum network (uncle blocks) after attack detection. dappctrl give ability to user to specify how many latest blocks are considered non reliable. These last blocks will not be used to fetch events or transactions.',
        'ethereum confirmation blocks');

INSERT INTO settings (key, value, description, name)
VALUES ('eth.event.maxretry',
        '7',
        'have value (stored as string) that is null or integer and represents how many times event should try to create job, until it is ignored. null or zero considered unlimited.',
        'event processing max retry');

INSERT INTO settings (key, value, description, name)
VALUES ('eth.event.freshofferings',
        '500',
        'defines number of latest block number to retrieve offerings for client. If eth.event.freshofferings is null or zero then all offerings will be downloaded.',
        'fresh offerings');

INSERT INTO settings (key, value, description, name)
VALUES ('error.sendremote',
        'true',
        'Allow error reporting to send logs to Privatix.',
        'error reporting');

END TRANSACTION;
