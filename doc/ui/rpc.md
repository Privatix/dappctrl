# UI JSON RPC

This document describes a UI JSON RPC API located in "ui" namespace.

## Synchronous methods

#### Accept Offering

*Method*:	`acceptOffering`

*Description*: Accept offering and create a new channel.

*Parameters*:
1. Password (string)
2. Account ethereum address (string)
3. Offering id (string)
4. Gas price (number)

*Result (object)*:
- `channel` (string) - id of channel to be created.

### Settings

#### Get Settings

*Method*:	`getSettings`

*Description*: Get settings.

*Parameters*:
1. Password (string)

*Result (object)*:
- `settings` (object) - object with keys as setting names and values as setting values.

#### Update Settings

*Method*:	`updateSettings`

*Description*: Update existing settings.

*Parameters*:
1. Password (string)
2. Object with keys as setting names and values as setting values (object)

*Result*: None.

#### Create Product

*Method*: `createProduct`

*Description*: Creates a new product.

*Parameters*: 
1. Password (string)
2. Product (`data.Product` object)

*Result (object)*:
- `product` (string) - id of created product.


#### Get Products

*Method*: `getProducts`

*Description*: Get all products available to the agent.

*Parameters*: None.

*Result (object)*:
1. Products (array of data.Product objects)


#### Set password

*Method*: `setPassword`

*Description*: Sets the password. Meant to be called only once. For password updates use `updatePassword`.

*Parameters*: 
1. Password (string)

*Result*: None.


#### Top Up Channel

*Method*: `topUpChannel`

*Description*: Top up a channel.

*Parameters*: 
1. Password (string)
2. Channel id (string)
3. Gas price (number)

*Result*: None.


#### Update Password

*Method*: `updatePassword`

*Description*: Updates the password.

*Parameters*: 
1. Current password (string)
2. New password (string)


#### Update Product

*Method*: `updateProduct`

*Description*: Updates a new product. If salt is 0, ignores its change. If password is empty, ignores its change.

*Parameters*: 
1. Password (string)
2. Product (`data.Product` object)

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

*Result (object)*:
- `object` (object) - object of a given type.

### Templates

#### Create Template

*Method*:	`createTemplate`

*Description*: Create new template.

*Parameters*:
1. Password (string)
2. Template (`data.Template` object)

*Result (object)*:
- `template` (string) - id of template to be created.

#### Get Templates

*Method*:	`getTemplates`

*Description*: Get templates.

*Parameters*:
1. Password (string)
2. Template type (string, can be `offer` or `access`)

*Result (object)*:
- `templates` (array of `data.Template` objects) - returned templates.

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
