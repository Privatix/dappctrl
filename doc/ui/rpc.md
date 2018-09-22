# UI JSON RPC

This document describes a UI JSON RPC API located in "ui" namespace.

## Synchronous methods

### Accounts

#### Export private key

*Method*:	`exportPrivateKey`

*Description*: Export a private key in Ethereum Keystore format by account id.

*Parameters*:
1. Password (string)
2. Account id (string)

*Result (array of `byte`s)*: private key in Ethereum Keystore format.

#### Generate Account

*Method*:	`generateAccount`

*Description*: Generate new private key and create new account.

*Parameters*:
1. Password (string)
2. Account (`data.Account` object)

*Result (string)*: id of account to be created.

#### Get accounts

*Method*:	`getAccounts`

*Description*: Get accounts.

*Parameters*:
1. Password (string)

*Result (array of `data.Account` objects)*: accounts.

#### Import Account From Hex

*Method*:	`importAccountFromHex`

*Description*: Import private key from hex and create new account.

*Parameters*:
1. Password (string)
2. Account (`data.Account` object)

*Result (string)*: id of account to be created.

#### Import Account From JSON

*Method*:	`importAccountFromJSON`

*Description*: Import private key from JSON blob with password and create new account.

*Parameters*:
1. Password (string)
2. Account (`data.Account` object)
3. Key in Ethereum keystore format (object)
4. Password to decrypting key in Ethereum keystore format (string)

*Result (string)*: id of account to be created.

#### Transfer tokens

*Method*:	`transferTokens`

*Description*: Create transfer of tokens between Privatix token & Privatix service contracts.

*Parameters*:
1. Password (string)
2. Account id (string)
3. Destination smart contract name (string, can be `ptc` or `psc`)
4. Token amount (number)
5. Gas price (number)

*Result*: None.

#### Update balance

*Method*:	`updateBalance`

*Description*: Actualize the PRIX token balance for a specific account.

*Parameters*:
1. Password (string)
2. Account id (string)

*Result*: None.


### Authentication

#### Set password

*Method*: `setPassword`

*Description*: Sets the password. Meant to be called only once. To update password use `updatePassword`.

*Parameters*: 
1. Password (string)

*Result*: None.

#### Update Password

*Method*: `updatePassword`

*Description*: Updates the password.

*Parameters*: 
1. Current password (string)
2. New password (string)

*Result*: None.


### Channels

#### Top Up Channel

*Method*: `topUpChannel`

*Description*: Top up a channel.

*Parameters*: 
1. Password (string)
2. Channel id (string)
3. Gas price (number)

*Result*: None.


### Offerings

#### Accept Offering

*Method*:	`acceptOffering`

*Description*: Accept offering and create a new channel.

*Parameters*:
1. Password (string)
2. Account ethereum address (string)
3. Offering id (string)
4. Deposit of tokens (number)
5. Gas price (number)

*Result (string)*: id of channel to be created.

#### Change Offering Status

*Method*:	`changeOfferingStatus`

*Description*: Change the status of a offering.

*Parameters*:
1. Password (string)
2. Offering id (string)
3. Action (string, can be `publish`, `popup` or `deactivate`)
4. Gas price (number)

*Result*: None.

#### Create Offering

*Method*:	`createOffering`

*Description*: Create offering.

*Parameters*:
1. Password (string)
2. Offering (`data.Offering` object)

*Result (string)*: id of offering to be created.

#### Get Offerings For Agent

*Method*:	`getAgentOfferings`

*Description*: Get active agent offerings.

*Parameters*:
1. Password (string)
2. Product id (string)
3. Offering status (string, can be `empty`, `register` or `remove`)

*Result (array of `data.Offering` objects)*: offerings.

#### Get Offerings For Client

*Method*:	`getClientOfferings`

*Description*: Get active client offerings.

*Parameters*:
1. Password (string)
2. Agent ethereum address (string)
3. Minimum unit price (number)
4. Maximum unit price (number)
5. Country codes ISO 3166-1 alpha-2 (array of strings)

*Result (array of `data.Offering` objects)*: offerings.

#### Update Offering

*Method*:	`updateOffering`

*Description*: Update an offering.

*Parameters*:
1. Password (string)
2. Offering (`data.Offering` object)

*Result*: None.


### Objects

#### Get object

*Method*:	`getObject`

*Description*: Get an object of a specified type..

*Parameters*:
1. Password (string)
2. Object type (string, can be `account`, `user`, `template`, `product`,
 `offering`, `channel`, `session`, `contract`, `endpoint`, `job`, `ethTx` or `ethLog`)
3. Object id (string)

*Result (object)*: object of a given type.


### Products

#### Create Product

*Method*: `createProduct`

*Description*: Creates a new product.

*Parameters*: 
1. Password (string)
2. Product (`data.Product` object)

*Result (string)*: id of created product.

#### Get Products

*Method*: `getProducts`

*Description*: Get all products available to the agent.

*Parameters*: None.

*Result (array of `data.Product` objects)*: products.

#### Update Product

*Method*: `updateProduct`

*Description*: Updates a new product. If salt is 0, ignores its change. If password is empty, ignores its change.

*Parameters*: 
1. Password (string)
2. Product (`data.Product` object)

*Result*: None.


### Sessions

#### Get Sessions

*Method*:	`getSessions`

*Description*: get sessions.

*Parameters*:
1. Password (string)
2. Channel id (string)

*Result (array of `data.Session` objects)*: sessions.


### Settings

#### Get Settings

*Method*:	`getSettings`

*Description*: Get settings.

*Parameters*:
1. Password (string)

*Result (object)*: object with keys as setting names and values as setting values.

#### Update Settings

*Method*:	`updateSettings`

*Description*: Update existing settings.

*Parameters*:
1. Password (string)
2. Object with keys as setting names and values as setting values (object)

*Result*: None.


### Templates

#### Create Template

*Method*:	`createTemplate`

*Description*: Create new template.

*Parameters*:
1. Password (string)
2. Template (`data.Template` object)

*Result (string)*: id of template to be created.

#### Get Templates

*Method*:	`getTemplates`

*Description*: Get templates.

*Parameters*:
1. Password (string)
2. Template type (string, can be `offer` or `access`)

*Result (array of `data.Template` objects)*: returned templates.


### Transactions

#### Get Ethereum Transactions

*Method*:	`getEthTransactions`

*Description*: Get Ethereum transactions.

*Parameters*:
1. Password (string)
2. Related type (string, can be `offering`, `channel`, `endpoint`, `account` or empty)
3. Related id (string, either uuid or empty)

*Result (array of `data.EthTx` objects)*: transactions.

## Subscriptions to asynchronous notifications

#### Object change

*Type*: `objectChange`

*Description*: Subscribe to changes for objects of a given type.

*Parameters*:
1. Password (string)
2. Type (string, can be `offering`, `channel`, `endpoint` or `account`)
3. Object ids (array of strings)

*Notification result (object)*:
- `object` (object) - changed object
- `job` (object) - job responsible for the change
- `error` (JSON RPC error object) - job error if it has failed
