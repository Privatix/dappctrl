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

*Result*:
- `channel` (string) - id of channel to be created.


#### Create Product

*Method*: `createProduct`

*Description*: Creates a new product.

*Parameters*: 
1. Password (string)
2. Product (`data.Product` object)

*Result*:
- `product` (string) - id of created product.


#### Get Products

*Method*: `getProducts`

*Description*: Get all products available to the agent.

*Parameters*: None.

*Result*:
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
