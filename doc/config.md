# dappctrl.config.json

## Fields description

### BlockMonitor
An ethereum blockchan monitor configuration

|Field|Type|Description|Example|
|-|-|-|-|
|QueryPause|int|Pause between iterations to query Ethereum logs in seconds|6|
|EthCallTimeout|int|Request timeout|5|

### DB
A database configuration

#### Connection
A database connection configuration

|Field|Type|Description|Example|
|-|-|-|-|
|user|string|A database user|postgres|
|dbname|string|A database name|dappctrl|
|sslmode|string|SSL mode|disable|
|port|number|postgres port|5432|

### Eth
An ethereum adapter configuration

|Field|Type|Description|Example|
|-|-|-|-|
|GethURL|string|Geth node URL|https://rinkeby.infura.io/k7mXdaE6eHJ4xMnOvx8Z|

#### Contract
An ethereum contracts configuration

|Field|Type|Description|Example|
|-|-|-|-|
|PTCAddr|string|Address of Privatix Token Contract|0xcA9a5951628486fAf8B9f58dB565E33ef9673394|
|PSCAddr|string|Address of Privatix Service Contract|0x10550c01b5c6f559d3dc78861400225ba88f3555|

### JobQueue
A job module configuration

|Field|Type|Description|Example|
|-|-|-|-|
|CollectJobs|uint|Number of jobs to process for collect-iteration|100|
|CollectPeriod|uint|Collect-iteration period, in milliseconds.|1000|
|WorkerBufLen|uint|Worker buffer length|10|
|Workers|uint|Number of workers, 0 means number of CPUs|0|
|TryLimit|uint8|Default number of tries to complete job|3|
|TryPeriod|uint|Default retry period, in milliseconds|60000|

#### Types
Job handlers overrides. Used to set custom parameters per job type.

|Field|Type|Description|Example|
|-|-|-|-|
|clientPreChannelCreate|struct|clientPreChannelCreate job settings|{"TryLimit": 3,"TryPeriod": 60000}|

### Log
A logger configuration.

|Field|Type|Description|Example|
|-|-|-|-|
|Level|string|Log levels: [debug, info, warning, error, fatal]|info|

### PayAddress

|||
|-|-|
|Type|string|
|Description|The address of payment endpoint on Agent, where Client will send cheques. Will be delivered to client via endpoint message|
|Example|http://localhost:9000/v1/pmtChannel/pay|

### PayServer
A payment server (listener) configuration

|Field|Type|Description|Example|
|-|-|-|-|
|Addr|string|Payment server address|localhost:9000|
|TLS|struct|Transport Layer Security settings|{"CertFile":"cert.pem","KeyFile": "key.pem",}|

### Role

|||
|-|-|
|Type|string|
|Description|The role of all users that run the application. Available choices are - agent,client.|
|Example|agent|
|Example|client|

### SessionServer
A session server configuration. Used to authorize and record service usage.

|Field|Type|Description|Example|
|-|-|-|-|
|Addr|string|Session server address|localhost:9000|
|TLS|struct|Transport Layer Security settings|{"CertFile":"cert.pem","KeyFile": "key.pem",}|

### SOMC
SOMC adapter configuration

|Field|Type|Description|Example|
|-|-|-|-|
|ReconnPeriod|int|Reconnection period in milliseconds|5000|
|URL|string|SOMC URL|ws://localhost:8080|

### SOMCServer
|Field|Type|Description|Example|
|-|-|-|-|
|Addr|int|the agents somc server address|5555|
|TLS|struct|Transport Layer Security settings| {"CertFile":"cert.pem","KeyFile": "key.pem"}| 

### SOMCServerAddr
An agent server (ui endpoint) configuration.

|Field|Type|Description|Example|
|-|-|-|-|
|Addr|string|The agent server ui endpoint address|localhost:3000|
|TLS|struct|Transport Layer Security settings| {"CertFile":"cert.pem","KeyFile": "key.pem"}|
|EthCallTimeout|uint|Ethereum operations call timeout in second|5|

### TorHostname

|||
|-|-|
|Type|string|
|Description|the agent hostname to send with offerings|
|Example|ssadfktgsdfsdfsdf.onion|

### TorSocksListener

|||
|-|-|
|Type|number|
|Description|Tor socks listener running port|
|Example|9050|

## Example

```
{
    "AgentServer": {
        "Addr": "localhost:3000",
        "TLS": null,
        "EthCallTimeout": 5
    },

    "BlockMonitor": {
        "CollectPause": 6,
        "SchedulePause": 6,
        "Timeout": 5
    },

    "DB": {
        "Conn": {
            "user": "postgres",
            "dbname": "dappctrl",
            "sslmode": "disable"
        }
    },

    "Eth": {
        "Contract" : {
            "PTCAddr": "",
            "PSCAddr": ""
        },
        "GethURL": ""
    },

    "JobQueue": {
        "CollectJobs": 100,
        "CollectPeriod": 1000,
        "WorkerBufLen": 10,
        "Workers": 0,
        "TryLimit": 3,
        "TryPeriod": 60000,
        "Types": {
            "clientPreChannelCreate": {
                "TryLimit": 3,
                "TryPeriod": 60000
            }
        }
    },

    "Log": {
        "Level": "info"
    },

    "PayAddress": "http://localhost:9000/v1/pmtChannel/pay",

    "PayServer": {
        "Addr": "localhost:9000",
        "TLS": null
    },

    "SessionServer": {
        "Addr": "localhost:8000",
        "TLS": null
    },

    "SOMC": {
        "ReconnPeriod": 5000,
        "URL": "ws://localhost:8080"
    }
}
```
