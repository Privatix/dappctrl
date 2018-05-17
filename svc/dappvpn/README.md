# OpenVPN Service Adapter

OpenVPN service adapter is an executable which integrates OpenVPN as a service
with the Privatix controller.

## Getting started

These instructions will help you to build and configure the OpenVPN service
adapter.

### Prerequisites

- Install OpenVPN 2.4+.
- Perform the Privatix controller installation steps described in the root
project `README.md`.

### Installation

Build the adapter:

```bash
go install $DAPPCTRL/svc/dappvpn
```

#### Additional steps for agent

Insert a new product into a database of the corresponding agent. Then modify
the adapter configuration file:

```bash
CONF_FILE=$DAPPCTRL_DIR/svc/dappvpn/dappvpn.config.json
LOCAL_CONF_FILE=$HOME/dappvpn.config.json
PRODUCT_ID=<uuid> # ID of a newly inserted product.
PRODUCT_PASS=<password> # Password of a newly inserted product.

jq ".Server.Username=\"$PRODUCT_ID\" | .Server.Password=\"$PRODUCT_PASS\"" $CONF_FILE > $LOCAL_CONF_FILE
```

Add the following lines to the `OpenVPN`-server configuration file
(substituting file paths):

```
auth-user-pass-verify "/path/to/dappvpn -config=/path/to/local/config" via-file
client-connect "/path/to/dappvpn -config=/path/to/local/config"
client-disconnect "/path/to/dappvpn -config=/path/to/local/config"
script-security 3
management localhost 7505
```

### Running the agent service

- Start the `OpenVPN`-server.
- Start the `dappvpn` in the background with the configuration provided.
