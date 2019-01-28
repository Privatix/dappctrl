# UI JSON RPC

This document describes a UI JSON RPC API located in "ui" namespace.

## Synchronous methods

### Accounts

#### Export private key

*Method*:	`exportPrivateKey`

*Description*: Export a private key in Ethereum Keystore format by account id.

*Parameters*:
1. Token (string)
2. Account id (string)

*Result (array of `byte`s)*: private key in Ethereum Keystore format.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_exportPrivateKey", "params": ["qwert", "3bc66565-9a8b-4b42-846d-0ae414065445"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": "eyJhZGRyZXNzIjoiOWEzNDNiYjMzMzczNDI4ZGM3MmQ0YWJmYWI0YzRhM2JlMzc2NWJlYyIsImNyeXB0byI6eyJjaXBoZXIiOiJhZXMtMTI4LWN0ciIsImNpcGhlcnRleHQiOiI2YjliODQ3OTdjM2E0OGEwYjFhYWJkZGUzNzZmODI2ZmQ5ZGJmOGQ1ODYyN2FiYWMzZjJmODFkZmNiZWI2Njc4IiwiY2lwaGVycGFyYW1zIjp7Iml2IjoiZjc4ZDRhYjI2YzJhYzU4NDRhNDFlYTNkODNiNzY1NzEifSwia2RmIjoic2NyeXB0Iiwia2RmcGFyYW1zIjp7ImRrbGVuIjozMiwibiI6MjYyMTQ0LCJwIjoxLCJyIjo4LCJzYWx0IjoiOGM5NDcxMDY2ZDJmNjIxNWY5YWMyZDEzYzhiMmM4MmM5MTg0NmM1MTUyNGIxYWY1MTFlOTYwMjVhNGYwOGFkZCJ9LCJtYWMiOiI1ZjQ4OGNhZDUyODZkYThhZWNkM2FlMDE3ODZlMDE4ZTRiNzc5MGZlODhkMGJkY2Q1YTQ1MTYwZGJkYmMyMmZjIn0sImlkIjoiOGMxMjg4MzYtOTcxNi00NTcyLWI3YjMtYzAzMDU4YWE5MTZmIiwidmVyc2lvbiI6M30="
}
```
</details>

#### Generate Account

*Method*:	`generateAccount`

*Description*: Generate new private key and create new account.

*Parameters*:
1. Token (string)
2. Account params (`ui.AccountParams` object)

*Result (string)*: id of account to be created.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_generateAccount", "params": ["qwert", {"name": "my acc", "isDefault": true, "inUse": true}], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": "b14a2e8e-fa08-4770-ba13-97d896a84980"
}
```
</details>

#### Get accounts

*Method*:	`getAccounts`

*Description*: Get accounts.

*Parameters*:
1. Token (string)

*Result (array of `data.Account` objects)*: accounts.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_generateAccount", "params": ["qwert", {"isDefault": true, "name": "my_acc"}], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": [
        {
            "id":"3bc66565-9a8b-4b42-846d-0ae414065445",
            "ethAddr":"9a343bb33373428dc72d4abfab4c4a3be3765bec",
            "isDefault":false,
            "inUse":false,
            "name":"my_acc",
            "ptcBalance":0,
            "pscBalance":0,
            "ethBalance":0,
            "lastBalanceCheck":null
        }
    ]
}
```
</details>

<details><summary>Example 2</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_generateAccount", "params": ["qwert", {"isDefault": true, "name": "my_acc"}], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": []
}
```
</details>

#### Import Account From Hex

*Method*:	`importAccountFromHex`

*Description*: Import private key from hex and create new account.

*Parameters*:
1. Token (string)
2. Account params with hex key (`ui.AccountParamsWithHexKey` object)

*Result (string)*: id of account to be created.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_importAccountFromHex", "params": ["qwert", {"isDefault": true, "name": "my_acc", "inUse": false, "privateKeyHex": "83ada09429152ff59ee8df29c687c3d96fc1c0bc9a9a703bb496a649e85dd9f3"}], "id": 67}' http://localhost:8888/http

// Result
{
    "id":6 7, 
    "jsonrpc": "2.0",
    "result": "8e0e455e-e11b-4341-95c3-1d66990eb22f"
}
```
</details>

#### Import Account From JSON

*Method*:	`importAccountFromJSON`

*Description*: Import private key from JSON blob with password and create new account.

*Parameters*:
1. Token (string)
2. Account params (`ui.AccountParams` object)
3. Key in Ethereum keystore format (object)
4. Password to decrypting key in Ethereum keystore format (string)

*Result (string)*: id of account to be created.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_importAccountFromJSON", "params": ["qwert", {"isDefault": true, "name": "acc from keystore", "inUse": true}, {"address":"4638140465c0ee8fc796323971431c30250433b2","crypto":{"cipher":"aes-128-ctr","ciphertext":"5d8749afaca5176b079d4b0ca96867ce2803795bb1edde1abb20c89a6d78a790","cipherparams":{"iv":"ba18922eae2d98291456dd5a2b38a7de"},"mac":"d3a288929127e36ba9edd191b2f48876f49290ad6bcd175592d6eb3180c13e2c","kdf":"pbkdf2","kdfparams":{"c":262144,"dklen":32,"prf":"hmac-sha256","salt":"a4d5a2ed2f65cee07309f966fe9d09c7ef16420f87e43cea5894029b3ee3e95c"}},"id":"8752ba1f-6930-4c87-acaa-766bda8f3ff1","version":3}, "qwerqwer"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": "e66d8abd-c5e4-4ced-b9c3-fc3d61a911d0"
}
```
</details>

#### Transfer tokens

*Method*:	`transferTokens`

*Description*: Create transfer of tokens between Privatix token & Privatix service contracts.

*Parameters*:
1. Token (string)
2. Account id (string)
3. Destination smart contract name (string, can be `ptc` or `psc`)
4. Token amount (number)
5. Gas price (number)

*Result*: None.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_transferTokens", "params": ["qwert", "e66d8abd-c5e4-4ced-b9c3-fc3d61a911d0", "pts", 10000000, 10000], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": null
}
```
</details>

#### Update account

*Method*:	`updateAccount`

*Description*: Updates an account.

*Parameters*:
1. Token (string)
2. Account id (string)
3. Name (string)
4. IsDefault (bool)
5. inUse (bool)

*Result*: None.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_updateAccount", "params": ["qwert", "e66d8abd-c5e4-4ced-b9c3-fc3d61a911d0", "user", true, true], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": null
}
```
</details>

#### Update balance

*Method*:	`updateBalance`

*Description*: Actualize the PRIX token balance for a specific account.

*Parameters*:
1. Token (string)
2. Account id (string)

*Result*: None.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_generateAccount", "params": ["qwert", "e66d8abd-c5e4-4ced-b9c3-fc3d61a911d0"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": null
}
```
</details>
</details>


### Authentication

#### Get Token

*Method*: `getToken`

*Description*: Given correct password, generates and returns new access token.

*Parameters*:
1. Password (string)

*Result*: Token (string).

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getToken", "params": ["qwert"], "id": 67}' http://localhost:8888/http

// Result
{
  "id": 67,
  "jsonrpc": "2.0",
  "result": "HyTT5U_u4WIXjuGM6-ruTF9_Zk1rKRtAB7BDhPWabtY=",
}
```
</details>

#### Set password

*Method*: `setPassword`

*Description*: Sets the password. Meant to be called only once. To update password use `updatePassword`.

*Parameters*: 
1. Password (string)

*Result*: None.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_setPassword", "params": ["qwert"], "id": 67}' http://localhost:8888/http

// Result
{
  "id": 67,
  "jsonrpc": "2.0",
  "result": null
}
```
</details>

#### Update Password

*Method*: `updatePassword`

*Description*: Updates the password.

*Parameters*: 
1. Current password (string)
2. New password (string)

*Result*: None.


<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_updatePassword", "params": ["qwert", "qwert2"], "id": 67}' http://localhost:8888/http

// Result
{
  "id": 67,
  "jsonrpc": "2.0",
  "result": null
}
```
</details>


### Channels

#### Change Channel Status

*Method*: `changeChannelStatus`

*Description*: Update channel state.

*Parameters*: 
1. Token (string)
2. Channel id (string)
3. Action (string, can be `terminate`, `pause`, `resume` or `close`)

*Result*: None.

<details><summary>Example</summary>

```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_changeChannelStatus", "params": ["qwerty","df78ff3e-666d-4b70-a158-240e6c655e8c","terminate"], "id": 67}' http://localhost:8888/http

// Result
{
  "id": 67,
  "jsonrpc": "2.0",
  "result": null
}
```

</details>

#### Get Agent Channels

*Method*: `getAgentChannels`

*Description*: Get channels for agent.

*Parameters*: 
1. Token (string)
2. Channel statuses (array of `string`s, can be empty)
3. Service statuses (array of `string`s, can be empty)
4. Offset (number)
5. Limit (number)

*Result (object)*:
- `items` (array of `data.Channel` objects) - channels.
- `totalItems` (number) - total channels.

<details><summary>Example</summary>

```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getAgentChannels", "params": ["qwerty",["active"],["pending", "terminated"],0,1], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[
            {
                "id":"aa79a640-540b-4a87-892c-7675b3f058b2",
                "agent":"829be313281957b2c7fb34c99c05b7b0affc09d3",
                "client":"7b8063c922db492c7116f79e84c5398c277f6d01",
                "offering":"c872f3bb-6e66-4fef-a70d-9d8e97542705",
                "block":1298498081,
                "channelStatus":"active",
                "serviceStatus":"pending",
                "serviceChangedTime":"2018-10-21T23:44:11.309Z",
                "preparedAt":"2018-11-21T23:44:11.309Z",
                "totalDeposit":100,
                "receiptBalance":5
            }
        ],
        "totalItems":70
    }
}
```

</details>

<details><summary>Example 2</summary>

```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getAgentChannels", "params": ["qwerty",[],[],0,1], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[
            {
                "id":"aa79a640-540b-4a87-892c-7675b3f058b2",
                "agent":"829be313281957b2c7fb34c99c05b7b0affc09d3",
                "client":"7b8063c922db492c7116f79e84c5398c277f6d01",
                "offering":"c872f3bb-6e66-4fef-a70d-9d8e97542705",
                "block":1298498081,
                "channelStatus":"active",
                "serviceStatus":"pending",
                "serviceChangedTime":"2018-10-21T23:44:11.309Z",
                "preparedAt":"2018-11-21T23:44:11.309Z",
                "totalDeposit":100,
                "receiptBalance":5
            }
        ],
        "totalItems":70
    }
}
```

</details>

<details><summary>Example 3</summary>

```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getAgentChannels", "params": ["qwerty",["active"],["pending"],0,1], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[],
        "totalItems":0
    }
}
```

</details>

#### Get Client Channels

*Method*: `getClientChannels`

*Description*: Get client channel information.

*Parameters*: 
1. Token (string)
2. Channel statuses (array of `string`s, can be empty)
3. Service statuses (array of `string`s, can be empty)
4. Offset (number)
5. Limit (number)

*Result (object)*:
- `items` (array of objects) - information of channels.
- `totalItems` (number) - total channel.

<details><summary>Example</summary>

```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getClientChannels", "params": ["qwerty",["active"],["pending", "terminated"],1,2], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[
            {
                "id":"df78ff3e-666d-4b70-a158-240e6c655e8c",
                "agent":"0x8D31cA7eBc9582874f15eac1caCa39A4782b3E06",
                "client":"0xC1bAE9F48e5cF5f16839F4BC1e312069003d7519",
                "offering":"a34bbecc-b294-4960-9a1c-bef468bd0617",
                "offeringHash":"tHC6By1U-m11YHwcCXTB3TdChp0SrJ28JuiYdBkEHMs=",
                "deposit":10000,
                "channelStatus":{
                    "serviceStatus":"pending",
                    "channelStatus":"active",
                    "lastChanged":"2018-10-21T23:44:11.309Z",
                    "maxInactiveTime":1800
                },
                "job":{
                    "id":"0cc65b67-29b9-4acb-bc44-0146caa7c6b4",
                    "jobtype":"clientPreChannelCreate",
                    "status":"done",
                    "createdAt":"2018-10-21T23:44:11.309Z"
                },
                "usage":{
                    "current":400,
                    "maxUsage":454,
                    "unitName":"MB",
                    "unitType":"units",
                    "cost":8811
                }
            }
        ],
        "totalItems":2
    }
}
```

</details>

<details><summary>Example 2</summary>

```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getClientChannels", "params": ["qwerty",[],[],1,2], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[
            {
                "id":"df78ff3e-666d-4b70-a158-240e6c655e8c",
                "agent":"0x8D31cA7eBc9582874f15eac1caCa39A4782b3E06",
                "client":"0xC1bAE9F48e5cF5f16839F4BC1e312069003d7519",
                "offering":"a34bbecc-b294-4960-9a1c-bef468bd0617",
                "offeringHash":"tHC6By1U-m11YHwcCXTB3TdChp0SrJ28JuiYdBkEHMs=",
                "deposit":10000,
                "channelStatus":{
                    "serviceStatus":"pending",
                    "channelStatus":"active",
                    "lastChanged":"2018-10-21T23:44:11.309Z",
                    "maxInactiveTime":1800
                },
                "job":{
                    "id":"0cc65b67-29b9-4acb-bc44-0146caa7c6b4",
                    "jobtype":"clientPreChannelCreate",
                    "status":"done",
                    "createdAt":"2018-10-21T23:44:11.309Z"
                },
                "usage":{
                    "current":400,
                    "maxUsage":454,
                    "unitName":"MB",
                    "unitType":"units",
                    "cost":8811
                }
            }
        ],
        "totalItems":2
    }
}
```

</details>

<details><summary>Example 3</summary>

```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getClientChannels", "params": ["qwerty",["active"],["pending"],1,2], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[],
        "totalItems":0
    }
}
```

</details>

#### Get Channels Usage

*Method*: `getChannelsUsage`

*Description*: Returns detailed usage of channels.


*Parameters*: 
1. Token (string)
2. Channels ids (array of `string`s)

*Result*:   []Usage (array of `ui.Usage` objects)

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getChannelsUsage", "params": ["qwert", ["e66d8abd-c5e4-4ced-b9c3-fc3d61a911d0"]], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": [{
                    "current":400,
                    "maxUsage":454,
                    "unitName":"MB",
                    "unitType":"units",
                    "cost":8811
                }]
}
```
</details>
</details>

#### Get Total Income

*Method*:   `getTotalIncome`

*Description*: Get total receipt balance from all channels.

*Parameters*: 
1. Token (string)

*Result*: Amount (number)

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getTotalIncome", "params": ["qwert"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": 12345
```
</details>
</details>

#### Top Up Channel

*Method*: `topUpChannel`

*Description*: Top up a channel.

*Parameters*: 
1. Token (string)
2. Channel id (string)
3. Deposit (number)
4. Gas price (number)

*Result*: None.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_topUpChannel", "params": ["qwert", "e66d8abd-c5e4-4ced-b9c3-fc3d61a911d0", 100000000, 10000], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": null
}
```
</details>
</details>


### Endpoints

#### Get Endpoints

*Method*:	`getEndpoints`

*Description*: get endpoints.

*Parameters*:
1. Token (string)
2. Channel id (string)
3. Template id (string)

*Result (array of `data.Endpoint` objects)*: endpoints.


### Objects

#### Get Object

*Method*:	`getObject`

*Description*: Get an object of a specified type.

*Parameters*:
1. Token (string)
2. Object type (string, can be `account`, `user`, `template`, `product`,
 `offering`, `channel`, `session`, `contract`, `endpoint`, `job`, `ethTx` or `ethLog`)
3. Object id (string)

*Result (object)*: object of a given type.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getObject", "params": ["qwert", "account", "e66d8abd-c5e4-4ced-b9c3-fc3d61a911d0"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": {
        "id":"e66d8abd-c5e4-4ced-b9c3-fc3d61a911d0",
        "ethAddr":"4638140465c0ee8fc796323971431c30250433b2",
        "isDefault":true,
        "inUse":true,
        "name":"acc from keystore",
        "ptcBalance":700000000,
        "pscBalance":0,
        "ethBalance":48085826000000000,
        "lastBalanceCheck":"2018-09-25T14:12:54.632205Z"
    }
}
```
</details>
</details>

#### Get Object By Hash

*Method*:	`getObjectByHash`

*Description*: Get an object of a specified type by hash.

*Parameters*:
1. Token (string)
2. Object type (string, can be `template`, `offering`, `endpoint`or `ethTx`)
3. Object hash (string)

*Result (object)*: object of a given type.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getObjectByHash", "params": ["qwert", "offering", "157df064ed3c2b555c0d670c9bcd744d9144915048ce0f61054395d5d98dfc"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": {
           "id":"687f26ab-5c62-4b05-8225-12e102a99450",
           "isLocal":false,
           "template":"efc61769-96c8-4c0d-b50a-e4d11fc30523",
           "product":"4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532",
           "hash":"157df064ed3c2b555c0d670c9bcd744d9144915048ce0f61054395d5d98dfc",
           "status":"empty",
           "blockNumberUpdated":1,
           "agent":"4638140465c0ee8fc796323971431c30250433b2",
           "rawMsg":"",
           "serviceName":"my service",
           "description":"my service description",
           "country":"KG",
           "supply":3,
           "currentSupply":3,
           "unitName":"",
           "unitType":"units",
           "billingType":"postpaid",
           "setupPrice":0,
           "unitPrice":100000,
           "minUnits":100,
           "maxUnit":null,
           "billingInterval":1800,
           "maxBillingUnitLag":1800,
           "maxSuspendTime":1800,
           "maxInactiveTimeSec":null,
           "freeUnits":0,
           "additionalParams":{},
           "autoPopUp":false
    }
}
```
</details>

### Offerings

#### Accept Offering

*Method*:	`acceptOffering`

*Description*: Accept offering and create a new channel.

*Parameters*:
1. Token (string)
2. Account ethereum address (string)
3. Offering id (string)
4. Deposit of tokens (number)
5. Gas price (number)

*Result (string)*: id of channel to be created.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_acceptOffering", "params": ["qwert", "4638140465c0ee8fc796323971431c30250433b2", "e66d8abd-c5e4-4ced-b9c3-fc3d61a911d0", 300000000, 10000], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": "7d9f3fb8-cd8c-43c0-af69-af59b879f3ad" 
}
```
</details>
</details>

#### Change Offering Status

*Method*:	`changeOfferingStatus`

*Description*: Change the status of a offering.

*Parameters*:
1. Token (string)
2. Offering id (string)
3. Action (string, can be `publish`, `popup` or `deactivate`)
4. Gas price (number)

*Result*: None.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_changeOfferingStatus", "params": ["qwert", "32989ae4-280b-4589-9062-632ba6217362", "popup", 10000], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": null
}
```
</details>
</details>

#### Create Offering

*Method*:	`createOffering`

*Description*: Create offering.

*Parameters*:
1. Token (string)
2. Offering (`data.Offering` object)

*Result (string)*: id of offering to be created.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_createOffering", "params": ["qwert", {"product": "4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532", "template": "efc61769-96c8-4c0d-b50a-e4d11fc30523", "agent": "0ba0e5f1-17f4-4f6d-b410-745a53048fc3", "serviceName": "my service", "description": "my service description", "country": "KG", "supply": 3, "unitName": "MB", "unitType": "units", "billingType": "postpaid", "setupPrice": 0, "unitPrice": 100000, "minUnits": 100, "maxUnit": 200, "billingInterval": 1, "maxBillingUnitLag": 3, "maxSuspendTime": 1800, "maxInactiveTimeSec": 1800, "freeUnits": 0, "additionalParams": {"minDownloadMbits":100,"minUploadMbits":80}, "autoPopUp":false}], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": "687f26ab-5c62-4b05-8225-12e102a99450"
}
```
</details>
</details>

#### Get Offerings For Agent

*Method*:	`getAgentOfferings`

*Description*: Get active agent offerings.

*Parameters*:
1. Token (string)
2. Product id (string)
3. Status (array of strings, each of which can be `empty`, `registering`, `registered`, `popping_up`, `popped_up`, `removing` or `removed`)
4. Offset (number)
5. Limit (number)

*Result (array of `data.Offering` objects)*: offerings.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getAgentOfferings", "params": ["qwert", "4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532", ["empty"], 0, 1], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[
            {
                "id":"687f26ab-5c62-4b05-8225-12e102a99450",
                "isLocal":false,
                "template":"efc61769-96c8-4c0d-b50a-e4d11fc30523",
                "product":"4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532",
                "hash":"157df064ed3c2b555c0d670c9bcd744d9144915048ce0f61054395d5d98dfc",
                "status":"empty",
                "blockNumberUpdated":1,
                "agent":"4638140465c0ee8fc796323971431c30250433b2",
                "rawMsg":"",
                "serviceName":"my service",
                "description":"my service description",
                "country":"KG",
                "supply":3,
                "currentSupply":3,
                "unitName":"",
                "unitType":"units",
                "billingType":"postpaid",
                "setupPrice":0,
                "unitPrice":100000,
                "minUnits":100,
                "maxUnit":null,
                "billingInterval":1800,
                "maxBillingUnitLag":1800,
                "maxSuspendTime":1800,
                "maxInactiveTimeSec":null,
                "freeUnits":0,
                "additionalParams":{},
                "autoPopUp":false
            }
        ],
        "totalItems":10
    }
}
```
</details>

<details><summary>Example 2</summary>

```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getAgentOfferings", "params": ["qwert", "4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532", "empty", 0, 1], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[],
        "totalItems":0
    }
}
```
</details>

#### Get Offerings For Client

*Method*:	`getClientOfferings`

*Description*: Get active client offerings.

*Parameters*:
1. Token (string)
2. Agent ethereum address (string)
3. Minimum unit price (number)
4. Maximum unit price (number)
5. Country codes ISO 3166-1 alpha-2 (array of strings)

*Result (array of `data.Offering` objects)*: offerings.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getClientOfferings", "params": ["qwert", "4638140465c0ee8fc796323971431c30250433b2", 0, 1000000, ["KG"], 0, 1], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[
            {
                "id":"687f26ab-5c62-4b05-8225-12e102a99450",
                "isLocal":false,
                "template":"efc61769-96c8-4c0d-b50a-e4d11fc30523",
                "product":"4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532",
                "hash":"                                            ",
                "status":"empty",
                "blockNumberUpdated":1,
                "agent":"4638140465c0ee8fc796323971431c30250433b2",
                "rawMsg":"",
                "serviceName":"my service",
                "description":"my service description",
                "country":"KG",
                "supply":3,
                "currentSupply":3,
                "unitName":"",
                "unitType":"units",
                "billingType":"postpaid",
                "setupPrice":0,
                "unitPrice":100000,
                "minUnits":100,
                "maxUnit":null,
                "billingInterval":1800,
                "maxBillingUnitLag":1800,
                "maxSuspendTime":1800,
                "maxInactiveTimeSec":null,
                "freeUnits":0,
                "additionalParams":{},
                "autoPopUp":null
            }
        ],
        "totalItems":10
    }
}
```
</details>

<details><summary>Example 2</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getClientOfferings", "params": ["qwert", "4638140465c0ee8fc796323971431c30250433b2", 0, 1000000, ["KG"], 0, 1], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[],
        "totalItems":0
    }
}
```

</details>

#### Get Offerings Filter Parameters For Client

*Method*:	`getClientOfferingsFilterParams`

*Description*: Get offerings filter parameters for client.

*Parameters*:
1. Token (string)

*Result (object)*:
- `countries` (array of strings) - Country codes ISO 3166-1 alpha-2.
- `minPrice` (number) - minimum value of minimal deposit required to accept the offering.
- `maxPrice` (number) - maximum value of minimal deposit required to accept the offering.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getClientOfferingsFilterParams", "params": ["qwerty"], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "countries":["SU","US"],
        "minPrice":121,
        "maxPrice":165
    }
}
```
</details>

<details><summary>Example 2</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getClientOfferingsFilterParams", "params": ["qwerty"], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "countries":[],
        "minPrice":0,
        "maxPrice":0
    }
}
```

</details>

#### Get Offering Income

*Method*: `getOfferingIncome`

*Description*: Get total receipt balance from all channels of offering with given id.

*Parameters*: 
1. Token (string)
2. Offering id (string)

*Result*: Amount (number)

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getOfferingIncome", "params": ["qwert", "687f26ab-5c62-4b05-8225-12e102a99450"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": 12345
}
```
</details>
</details>

#### Get Offering Usage

*Method*: `getOfferingUsage`

*Description*: Returns total units used for all channels with a given offering.

*Parameters*: 
1. Token (string)
2. Offering id (string)

*Result*:   Amount (number)

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getOfferingUsage", "params": ["qwert", "687f26ab-5c62-4b05-8225-12e102a99450"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": 12345
}
```
</details>
</details>

#### Update Offering

*Method*:	`updateOffering`

*Description*: Update an offering.

*Parameters*:
1. Token (string)
2. Offering (`data.Offering` object)

*Result*: None.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_updateOffering", "params": ["qwert", {"id":"687f26ab-5c62-4b05-8225-12e102a99450","isLocal":false,"template":"efc61769-96c8-4c0d-b50a-e4d11fc30523","product":"4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532","hash":"                                            ","status":"empty","blockNumberUpdated":1,"agent":"4638140465c0ee8fc796323971431c30250433b2","rawMsg":"","serviceName":"my service 2","description":"my service description 2","country":"KG","supply":3,"currentSupply":3,"unitName":"","unitType":"units","billingType":"postpaid","setupPrice":0,"unitPrice":100000,"minUnits":100,"maxUnit":null,"billingInterval":1800,"maxBillingUnitLag":1800,"maxSuspendTime":1800,"maxInactiveTimeSec":null,"freeUnits":0,"additionalParams":{},"autoPopUp":true}], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": null
}
```
</details>
</details>

#### Ping Offerings

*Method*:	`pingOfferings`

*Description*: Ping offerings.

*Parameters*:
1. Token (string)
2. IDs of Offerings to ping (string)

*Result*: Object with IDs of Offerings as keys as ping result as values.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_pingOfferings", "params": ["qwert", ["687f26ab-5c62-4b05-8225-12e102a99450"]], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": {
        "687f26ab-5c62-4b05-8225-12e102a99450": true
    }
}
```
</details>
</details>


### Ethereum Logs

#### Get Last Block Number

*Method*:   `getLastBlockNumber`

*Description*: returns max(block_number) of collected ethereum logs + min confirmations setting value.

*Parameters*:
1. Token (string)

*Result*: Block Number (number)

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getLastBlockNumber", "params": ["qwert"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": 1234
}
```
</details>
</details>

### Logs

#### Get Logs

*Method*:   `getLogs`

*Description*: Get back end log, paginated.

*Parameters*:
1. Token (string)
2. Log levels (array of strings, strings can be `debug`, `info`, `warning`, `error` or `fatal`) 
3. Search text (string)
4. Lower bound of the filter by time. Time in ISO 8601 RFC 3339 format (string)
5. Upper bound of the filter by time. Time in ISO 8601 RFC 3339 format (string)
6. Offset (number)
7. Limit (number)

*Result (object)*:
- `items` (array of `data.LogEvent` objects) - log events.
- `totalItems` (number) - total items.

<details><summary>Example1</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getLogs", "params": ["qwerty", ["error", "warning"], "","2018-10-19T10:47:22","2019-10-19T10:47:22", 0, 2], "id": 67}' http://localhost:8888/http

// Result
{
    {
        "jsonrpc":"2.0",
        "id":67,
        "result":{
            "items":[
                {
                    "time":"2018-10-21T14:56:42.718817+07:00",
                    "level":"error",
                    "message":"key eth.min.confirmations is not exist in Setting table",
                    "context":{
                        "type":"monitor.Monitor",
                        "method":"Start"
                    },
                    "stack":"goroutine 8 [running]:\nruntime/debug.Stack(0x9db8e0, 0xc420293830, 0xc420409880)\n\t/usr/local/go/src/runtime/debug/stack.go:24 +0xa7\ngithub.com/privatix/dappctrl/util/log.(*LoggerBase).Log(0xc420443740, 0xad27bf, 0x5, 0xc4200dc700, 0x37)\n\t/home/bik/go/src/github.com/privatix/dappctrl/util/log/logger.go:113 +0x1be\ngithub.com/privatix/dappctrl/util/log.multiLogger.Log(0xc420174c00, 0x4, 0x4, 0xad27bf, 0x5, 0xc4200dc700, 0x37)\n\t/home/bik/go/src/github.com/privatix/dappctrl/util/log/multi_logger.go:20 +0x69\ngithub.com/privatix/dappctrl/util/log.multiLogger.Error(0xc420174c00, 0x4, 0x4, 0xc4200dc700, 0x37)\n\t/home/bik/go/src/github.com/privatix/dappctrl/util/log/multi_logger.go:27 +0x68\ngithub.com/privatix/dappctrl/monitor.(*Monitor).Start.func2(0xb88de0, 0xc420174c80, 0xc420508460, 0xc42028a380, 0xb8bd20, 0xc420443780)\n\t/home/bik/go/src/github.com/privatix/dappctrl/monitor/monitor.go:119 +0x20b\ncreated by github.com/privatix/dappctrl/monitor.(*Monitor).Start\n\t/home/bik/go/src/github.com/privatix/dappctrl/monitor/monitor.go:110 +0x1c2\n"
                }
            ],
            "totalItems":33
        }
    }
}
```

<details><summary>Example2</summary>

```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getLogs", "params": ["qwerty", [], "monitor.Monitor","2018-10-19T10:47:22","2019-10-19T10:47:22", 0, 1], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[
            {
                "time":"2018-10-21T15:00:40.286333+07:00",
                "level":"error",
                "message":"key eth.min.confirmations is not exist in Setting table",
                "context":{
                    "type":"monitor.Monitor",
                    "method":"Start"
                },
                "stack":"goroutine 11 [running]:\nruntime/debug.Stack(0x9db880, 0xc420299360, 0xc42040f860)\n\t/usr/local/go/src/runtime/debug/stack.go:24 +0xa7\ngithub.com/privatix/dappctrl/util/log.(*LoggerBase).Log(0xc4204db260, 0xad275f, 0x5, 0xc420211080, 0x37)\n\t/home/bik/go/src/github.com/privatix/dappctrl/util/log/logger.go:113 +0x1be\ngithub.com/privatix/dappctrl/util/log.multiLogger.Log(0xc420158c80, 0x4, 0x4, 0xad275f, 0x5, 0xc420211080, 0x37)\n\t/home/bik/go/src/github.com/privatix/dappctrl/util/log/multi_logger.go:20 +0x69\ngithub.com/privatix/dappctrl/util/log.multiLogger.Error(0xc420158c80, 0x4, 0x4, 0xc420211080, 0x37)\n\t/home/bik/go/src/github.com/privatix/dappctrl/util/log/multi_logger.go:27 +0x68\ngithub.com/privatix/dappctrl/monitor.(*Monitor).Start.func2(0xb88d80, 0xc420158cc0, 0xc420170460, 0xc420290380, 0xb8bcc0, 0xc4204db2a0)\n\t/home/bik/go/src/github.com/privatix/dappctrl/monitor/monitor.go:119 +0x20b\ncreated by github.com/privatix/dappctrl/monitor.(*Monitor).Start\n\t/home/bik/go/src/github.com/privatix/dappctrl/monitor/monitor.go:110 +0x1c2\n"
            }
        ],
        "totalItems":110
    }
}
```

</details>

<details><summary>Example 3</summary>

```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getLogs", "params": ["qwerty", 0,1,"monitor.Monitor","","",""], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[],
        "totalItems":0
    }
}
```

</details>

### Products

#### Create Product

*Method*: `createProduct`

*Description*: Creates a new product.

*Parameters*: 
1. Token (string)
2. Product (`data.Product` object)

*Result (string)*: id of created product.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_createProduct", "params": ["qwert", {"name": "my product", "offerTplID": "efc61769-96c8-4c0d-b50a-e4d11fc30523", "offerAccessID": "d0dfbbb2-dd07-423a-8ce0-1e74ce50105b", "usageRepType": "total", "isServer": true, "clientIdent": "by_channel_id", "config": {}}], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": "35d5ed75-7677-43b7-aa94-19eba10c6f23"
}
```
</details>
</details>

#### Get Products

*Method*: `getProducts`

*Description*: Get all products available to the agent.

*Parameters*:
1. Token (string)

*Result (array of `data.Product` objects)*: products.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getProducts", "params": ["qwert"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": [
        {
            "id":"4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532",
            "name":"VPN server",
            "offerTplID":"efc61769-96c8-4c0d-b50a-e4d11fc30523",
            "offerAccessID":"d0dfbbb2-dd07-423a-8ce0-1e74ce50105b",
            "usageRepType":"total",
            "isServer":true,
            "clientIdent":"by_channel_id",
            "config":{"somekey":"somevalue"},
            "serviceEndpointAddress":"127.0.0.1"
        },
        {
            "id":"35d5ed75-7677-43b7-aa94-19eba10c6f23",
            "name":"my product",
            "offerTplID":"efc61769-96c8-4c0d-b50a-e4d11fc30523",
            "offerAccessID":"d0dfbbb2-dd07-423a-8ce0-1e74ce50105b",
            "usageRepType":"total",
            "isServer":true,
            "clientIdent":"by_channel_id",
            "config":{},
            "serviceEndpointAddress":null
        }
    ]
}
```
</details>

<details><summary>Example 2</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getProducts", "params": ["qwert"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": []
}
```
</details>

#### Get Product Income

*Method*:   `getProductIncome`

*Description*: Get total receipt balance from all channels of all offerings with given product id.

*Parameters*: 
1. Token (string)
2. Product id (string)

*Result*: Amount (number)

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getProductIncome", "params": ["qwert", "35d5ed75-7677-43b7-aa94-19eba10c6f23"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": 12345
```
</details>
</details>

#### Get Product usage

*Method*: `getProductUsage`

*Description*: Returns total units used in all channel of all offerings with a given product.

*Parameters*: 
1. Token (string)
2. Product id (string)

*Result*:   Amount (number)

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getProductUsage", "params": ["qwert", "35d5ed75-7677-43b7-aa94-19eba10c6f23"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": 12345
}
```
</details>
</details>

#### Update Product

*Method*: `updateProduct`

*Description*: Updates a new product. If salt is 0, ignores its change. If password is empty, ignores its change.

*Parameters*: 
1. Token (string)
2. Product (`data.Product` object)

*Result*: None.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_updateProduct", "params": ["qwert",{"id":"35d5ed75-7677-43b7-aa94-19eba10c6f23","name":"my product","offerTplID":"efc61769-96c8-4c0d-b50a-e4d11fc30523","offerAccessID":"d0dfbbb2-dd07-423a-8ce0-1e74ce50105b","usageRepType":"total","isServer":true,"clientIdent":"by_channel_id","config":{}}], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": null
}
```
</details>
</details>


### Sessions

#### Get Sessions

*Method*:	`getSessions`

*Description*: get sessions.

*Parameters*:
1. Token (string)
2. Channel id (string)

*Result (array of `data.Session` objects)*: sessions.


### Settings

#### Get Settings

*Method*:	`getSettings`

*Description*: Get settings.

*Parameters*:
1. Token (string)

*Result (object)*: object with keys as setting keys and values as setting information.
- `value` (string) - setting value.
- `permissions` (string, can be `readWrite` or `readOnly`) - setting permissions.

<details><summary>Example 1</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getSettings", "params": ["qwert"], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "error.sendremote":{
            "value":"true",
            "permissions":"readWrite"
        },
        "eth.default.gasprice":{
            "value":"20000000000",
            "permissions":"readWrite"
        },
        "eth.event.blocklimit":{
            "value":"80",
            "permissions":"readWrite"
        },
        "eth.event.freshblocks":{
            "value":"11520",
            "permissions":"readWrite"
        },
        "eth.event.lastProcessedBlock":{
            "value":"3263580",
            "permissions":"readWrite"
        },
        "eth.max.deposit":{
            "value":"30000000000",
            "permissions":"readWrite"
        },
        "eth.min.confirmations":{
            "value":"1",
            "permissions":"readWrite"
        },
        "offering.autopopup":{
            "value":"true",
            "permissions":"readWrite"
        },
        "system.version.db":{
            "value":"0.14.0",
            "permissions":"readOnly"
        }
    }
}

```
</details>

<details><summary>Example 2</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getSettings", "params": ["qwert"], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{}
}

```
</details>

#### Update Settings

*Method*:	`updateSettings`

*Description*: Update existing settings.

*Parameters*:
1. Token (string)
2. Object with keys as setting names and values as setting values (object)

*Result*: None.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_updateSettings", "params": ["qwert", {"eth.default.gasprice":"12345"}], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": null
}
```
</details>
</details>


### Settings GUI

#### GetGUISettings

*Method*:   `getGUISettings`

*Description*: Get GUI settings.


*Parameters*:
1. Token (string)

*Result*: js object.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getGUISettings", "params": ["qwert"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": {"foo": "bar"}
}
```
</details>
</details>

#### SetGUISettings

*Method*:   `setGUISettings`

*Description*: Set GUI settings.


*Parameters*:
1. Token (string)
2. JS object.

*Result*: None.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_setGUISettings", "params": ["qwert", {"foo": "bar"}], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": null
}
```
</details>
</details>


### Templates

#### Create Template

*Method*:	`createTemplate`

*Description*: Create new template.

*Parameters*:
1. Token (string)
2. Template (`data.Template` object)

*Result (string)*: id of template to be created.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_createTemplate", "params": ["qwert", {"hash": "", "raw": {}, "kind": "offer"}], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": "457a9e1c-9236-4713-86ab-73cf3a7c86c5"
}
```
</details>
</details>

#### Get Templates

*Method*:	`getTemplates`

*Description*: Get templates.

*Parameters*:
1. Token (string)
2. Template type (string, can be `offer` or `access`)

*Result (array of `data.Template` objects)*: returned templates.

<details><summary>Example</summary>
    
```js
// Request
curl -X GET -H "Content-Type: application/json" --data '{"method": "ui_getTemplates", "params": ["qwert", "access"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": [
        {
            "id":"d0dfbbb2-dd07-423a-8ce0-1e74ce50105b",
            "hash":"157df064ed3c2b555c0d670c9bcd744d9144915048ce0f61054395d5d98dfc",
            "raw":{"definitions":{"host":{"pattern":"^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])(\\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9]))*:[0-9]{2,5}$","type":"string"},
            "simple_url":{"pattern":"^(http:\\/\\/www\\.|https:\\/\\/www\\.|http:\\/\\/|https:\\/\\/)?.+","type":"string"},
            "uuid":{"pattern":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}","type":"string"}},
            "properties":{"additionalParams":{"additionalProperties":{"type":"string"},
            "minProperties":1,"type":"object"},
            "password":{"type":"string"},
            "paymentReceiverAddress":{"$ref":"#/definitions/simple_url"},
            "serviceEndpointAddress":{"type":"string"},
            "templateHash":{"type":"string"},
            "username":{"$ref":"#/definitions/uuid"}},
            "required":["templateHash","paymentReceiverAddress","serviceEndpointAddress","additionalParams"],
            "title":"Privatix VPN access",
            "type":"object"
            },
            "kind":"access"
        }
    ]
}
```
</details>

<details><summary>Example 2</summary>
    
```js
// Request
curl -X GET -H "Content-Type: application/json" --data '{"method": "ui_getTemplates", "params": ["qwert", "access"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": []
}
```
</details>


### Transactions

#### Get Ethereum Transactions

*Method*:	`getEthTransactions`

*Description*: Get Ethereum transactions. If related type is `accountAggregated`, then get an Ethereum address of the account and find all transactions where this address is the sender.

*Parameters*:
1. Token (string)
2. Related type (string, can be `offering`, `channel`, `endpoint`, `account`, `accountAggregated` or empty)
3. Related id (string, either uuid or empty)
4. Offset (number)
5. Limit (number)

*Result (array of `data.EthTx` objects)*: transactions.

<details><summary>Example</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getEthTransactions", "params": ["qwert", "channel", "d0dfbbb2-dd07-423a-8ce0-1e74ce50105b", 0, 1], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[
            {
                "id": "bc8310c8-6709-4cb0-9642-703f2bd3bb5d",
                "hash": "36845db5d50bd1ac0ea31353a20b8a9616279dcf000b51eca995debca678c94c",
                "method": "CreateChannel",
                "status": "sent",
                "job": "dd123187-ff00-4785-ac1f-f3f6fb28ac35",
                "issued": "2018-09-18 10:01:22.055041+02",
                "addrFrom": "e4b2ad904ab4b4e70c58c0beb04d6e46522b2858",
                "addrTo": "0381ce1568a3219b0bf8f4126939322cf7248510",
                "nonce": 168,
                "gasPrice": 6000000000,
                "gas": 200000,
                "txRaw": {"r": "0x622029910b2949d4011df8d615744c4ce247629162f629ea38410749a10cb3e5", "s": "0x13215fa2b64a15cdc5371cff9baa1995ab8b7220c7117dbfedb520bf119a30a5", "v": "0x1b", "to": "0xae6bfd07c02b1fca7e1cbc160a87729f3fafb794", "gas": "0x30d40", "hash": "0x36845db5d50bd1ac0ea31353a20b8a9616279dcf000b51eca995debca678c94c", "input": "0x6bc371520000000000000000000000000381ce1568a3219b0bf8f4126939322cf7248510e6a24e1e28d3c2573db24fb07aaebe8aad05e08342bf7c8661d0ad7860acf04000000000000000000000000000000000000000000000000000000000007a12000000000000000000000000000000000000000000000000000000000000000000", "nonce": "0xa8", "value": "0x0", "gasPrice": "0x165a0bc00"},
                "relatedType": "channel",
                "relatedID": "1e9417f5-dea7-4944-9a8e-4b9a002c0c72",
            }
        ],
        "totalItems":10
    }
}
```
</details>

<details><summary>Example 2</summary>
    
```js
// Request
curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getEthTransactions", "params": ["qwert", "channel", "d0dfbbb2-dd07-423a-8ce0-1e74ce50105b, 0, 1], "id": 67}' http://localhost:8888/http

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":{
        "items":[],
        "totalItems":0
    }
}
```
</details>

#### Get User Role

*Method*:	`getUserRole`

*Description*: Get user role.

*Parameters*:
1. Token (string)

*Result (string)*: user role, can be `agent` or `client`.

<details><summary>Example</summary>
    
```js
// Request
curl -X GET -H "Content-Type: application/json" --data '{"method": "ui_getUserRole", "params": ["qwert"], "id": 67}' http://localhost:8888/http

// Result
{
    "id": 67,
    "jsonrpc": "2.0",
    "result": "client"
}
```
</details>

## Subscriptions to asynchronous notifications

#### Object change

*Type*: `objectChange`

*Description*: Subscribe to changes for objects of a given type.

*Parameters*:
1. Token (string)
2. Type (string, can be `offering`, `channel`, `endpoint` or `account`)
3. Subscription keys (each key can be a related object id or a job type, array
of strings)

*Notification result (object)*:
- `object` (object) - changed object
- `job` (object) - job responsible for the change
- `error` (JSON RPC error object) - job error if it has failed
