# dappctrl.config.json

## Fields description

### AgentMonitor
Monitors payments and usage for services and trigger service status transition if needed.

|Field|Type|Description|Example|
|-|-|-|-|
|Interval|uint64|Interval between round checks in milliseconds|5000|

### BlockMonitor
An ethereum blockchan monitor configuration

|Field|Type|Description|Example|
|-|-|-|-|
|QueryPause|int|Pause between iterations to query Ethereum logs in seconds|6|
|EthCallTimeout|int|Request timeout|5|

### ClientMonitor
Monitors billing for active client channels.

|Field|Type|Description|Example|
|-|-|-|-|
|CollectPeriod|uint|Period between rounds in milliseconds|5000|
|RequestTLS|bool|Wether use https or not on payments sending|false|
|RequestTimeout|uint|In milliseconds, must be less than CollectPeriod|2500|


### Country
Desription

|Field|Type|Description|Example|
|-|-|-|-|
|Field|string|Field to extract country code from|country_code|
|Timeout|uint64|Country retrieve request timeout in milliseconds|30000|
|URLTemplate|string|Address to retrieve country details by ip|https://country.example.com/{{ip}}|

### DB
A database configuration

#### Conn
A database connection configuration

|Field|Type|Description|Example|
|-|-|-|-|
|user|string|A database user|postgres|
|dbname|string|A database name|dappctrl|
|sslmode|string|SSL mode|disable|
|port|number|postgres port|5432|

#### MaxOpen
The maximum number of open connections to the database.

#### MaxIddle
The maximum number of connections in the idle connection pool.

### DBLog
Database logger configuration.

|Field|Type|Description|Example|
|-|-|-|-|
|Level|string||info|
|StackLevel|string||error|

### EptMsg
Desription

|Field|Type|Description|Example|
|-|-|-|-|
|Timeout|uint|Timeout to generate an endpoint message in milliseconds|1000|

### Eth
An ethereum adapter configuration

|Field|Type|Description|Example|
|-|-|-|-|
|CheckTimeout|uint64|Period to check connection status in milliseconds|10000|
|GethURL|string|Geth node URL|https://rinkeby.infura.io/k7mXdaE6eHJ4xMnOvx8Z|
|Timeout|uint64|Request timeout to GethURL in milliseconds|120000|

#### Contract
An ethereum contracts configuration

|Field|Type|Description|Example|
|-|-|-|-|
|PTCAddr|string|Address of Privatix Token Contract|0xcA9a5951628486fAf8B9f58dB565E33ef9673394|
|PSCAddr|string|Address of Privatix Service Contract|0x10550c01b5c6f559d3dc78861400225ba88f3555|

##### Periods
An ethereum contracts configuration

|Field|Type|Description|Example|
|-|-|-|-|
|Challenge|uint32|Challange period in number of blocks|20|
|PopUp|uint32|PopUp period in number of blocks|10|
|Remove|uint32|Remove period in number of blocks|5|

#### HTTPClient
sdfsdf

|Field|Type|Description|Example|
|DialTimeout|uint64||10000|
|IdleConnTimeout|uint64||10000|
|KeepAliveTimeout|uint64||10000|
|RequestTimeout|uint64||10000|
|ResponseHeaderTimeout|uint64||10000|
|TLSHandshakeTimeout|uint64||10000|

### FileLog
Desription

|Field|Type|Description|Example|
|-|-|-|-|
|FileMode|uint32|Mode to create log file with|0644|
|Filename|string|Path to the file with date parts like %Y %m %d|/var/log/dappctrl-%Y-%m-%d.log|
|Level|string||info|
|Prefix|string|Prefix to write at beginning of each line|dappctrl-|
|StackLevel|string||debug|
|UTC|bool|Whether to use UTC or not for time logging|false|

### Gas
Default gas limits for contract calls.

#### PSC
Default gas limits for psc methods.

|Field|Type|Description|Example|
|-|-|-|-|
|AddBalanceERC20|uint64||100000|
|CooperativeClose|uint64||100000|
|CreateChannel|uint64||100000|
|PopupServiceOffering|uint64||100000|
|RegisterServiceOffering|uint64||100000|
|RemoveServiceOffering|uint64||100000|
|ReturnBalanceERC20|uint64||100000|
|SetNetworkFee|uint64||100000|
|Settle|uint64||100000|
|TopUp|uint64||100000|
|UncooperativeClose|uint64||100000|

#### PTC
Default gas limits for ptc methods.

|Field|Type|Description|Example|
|-|-|-|-|
|Approve|uint64||100000|

### Job
A job module configuration

|Field|Type|Description|Example|
|-|-|-|-|
|CollectJobs|uint|Number of jobs to process for collect-iteration|100|
|CollectPeriod|uint|Collect-iteration period, in milliseconds.|1000|
|TryLimit|uint8|Default number of tries to complete job|3|
|TryPeriod|uint|Default retry period, in milliseconds|60000|
|WorkerBufLen|uint|Worker buffer length|10|
|Workers|uint|Number of workers, 0 means number of CPUs|0|

#### Types
Job handlers overrides. Used to set custom parameters per job type.

|Field|Type|Description|Example|
|-|-|-|-|
|clientPreChannelCreate|struct|clientPreChannelCreate job settings|{"TryLimit": 3,"TryPeriod": 60000}|

### Looper
Desription

|Field|Type|Description|Example|
|-|-|-|-|
|AutoOfferingPopUpTimeout|uint64|Period duration between offerings auto pop ups in milliseconds|3600000|

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

### Report
Bugsnag configuration.

|Field|Type|Description|Example|
|-|-|-|-|
|AppID|string|||
|ReleaseStage|string|||
|ExcludedPackages|[]string|||

### ReportLog
Desription

|Field|Type|Description|Example|
|-|-|-|-|
|Level|string||info|
|StackLevel|string||error|

### Role
Either "client" or "agent".

### SOMCServer
Independent agent somc server. Intended to be shared via tor net. For agents only.

|Field|Type|Description|Example|
|-|-|-|-|
|Addr|int|the agents somc server address|5555|
|TLS|struct|Transport Layer Security settings| {"CertFile":"cert.pem","KeyFile": "key.pem"}| 

### SessionServer
A session server configuration. Used to authorize and record service usage.

|Field|Type|Description|Example|
|-|-|-|-|
|Addr|string|Session server address|localhost:9000|
|TLS|struct|Transport Layer Security settings|{"CertFile":"cert.pem","KeyFile": "key.pem",}|

### StaticPassword
If specified, uses this password for authentication and encryption.

### TorHostname
Hostname to publish with offerings as somc endpoint. For agents only.

### TorSocksListener
Port used to connect to tor net. For clients only.

### UI
UI api server configuration.

|Field|Type|Description|Example|
|-|-|-|-|
|Addr|string|Session server address|localhost:9000|
|AllowedOrigins|string||["*"]|
|TLS|struct|Transport Layer Security settings|{"CertFile":"cert.pem","KeyFile": "key.pem",}|


## Example

```
{
    "AgentMonitor": {
        "Interval": 5000
    },
    "BlockMonitor": {
        "QueryPause": 6000,
        "EthCallTimeout": 60000
    },
    "ClientMonitor": {
        "CollectPeriod": 5000,
        "RequestTLS": false,
        "RequestTimeout": 2500
    },
    "Country": {
        "Field" : "country_code",
        "Timeout": 30,
        "URLTemplate" : "https:/country.example.com/{{ip}}"
    },
    "DB": {
        "Conn": {
            "dbname": "dappctrl",
            "sslmode": "disable",
            "user": "postgres",
            "host": "localhost",
            "port": "5432"
        }
    },
    "DBLog": {
        "Level": "info",
        "StackLevel": "error"
    },
    "EptMsg": {
        "Timeout": 1
    },
    "Eth": {
        "CheckTimeout": 10000,
        "Contract": {
            "PTCAddrHex": "0x0d825eb81b996c67a55f7da350b6e73bab3cb0ec",
            "PSCAddrHex": "0xef85cd1c955b36945687d56b1cec824ad84bd684",
            "Periods": {
                "PopUp": 10,
                "Challenge": 20,
                "Remove": 5
            }
        },
        "GethURL": "https://rinkeby.infura.io/v3/6396832f7ea1488ba30fb4de8f6b06ea",
        "Timeout": 120,
        "HTTPClient": {
            "DialTimeout": 10,
            "TLSHandshakeTimeout": 5,
            "ResponseHeaderTimeout": 60,
            "RequestTimeout": 60,
            "IdleConnTimeout": 60,
            "KeepAliveTimeout": 60
        }
    },
    "FileLog": {
        "Level": "debug",
        "StackLevel": "error",
        "Prefix": "",
        "UTC": false,
        "Filename": "/var/log/dappctrl-%Y-%m-%d.log",
        "FileMode": 420
    },
    "Gas": {
        "PTC": {
            "Approve": 100000
        },
        "PSC": {
            "AddBalanceERC20": 100000,
            "RegisterServiceOffering": 200000,
            "CreateChannel": 200000,
            "CooperativeClose": 200000,
            "ReturnBalanceERC20": 100000,
            "SetNetworkFee": 100000,
            "UncooperativeClose": 100000,
            "Settle": 100000,
            "TopUp": 100000,
            "PopupServiceOffering": 100000,
            "RemoveServiceOffering": 100000
        }
    },
    "Job": {
        "CollectJobs": 100,
        "CollectPeriod": 1000,
        "WorkerBufLen": 10,
        "Workers": 0,
        "TryLimit": 3,
        "TryPeriod": 60000,
        "Types": {
            "clientAfterOfferingMsgBCPublish": {
                "TryLimit": 10,
                "TryPeriod": 60000,
                "FirstStartDelay": 30000
            }
        }
    },
    "Looper": {
        "AutoOfferingPopUpTimeout": 3600000
    },
    "PayAddress": "http://0.0.0.0:9000/v1/pmtChannel/pay",
    "PayServer": {
        "Addr": "0.0.0.0:9000",
        "TLS": null
    },
    "Role": "agent",
    "SessionServer": {
        "Addr": "localhost:8000",
        "TLS": null
    },
    "SOMC": {
        "URL": "ws://89.38.96.53:8080"
    },
    "SOMCServer": {
        "Addr": "localhost:5555",
        "TLS": null
    },
    "StaticPassword": "",
    "TorHostname": "",
    "TorSocksListener": 9050,
    "UI": {
        "Addr": "localhost:8888",
        "AllowedOrigins": ["*"],
        "TLS": null
    }
}

```
