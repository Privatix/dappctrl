# JSON RPC

This document describes a generic JSON RPC API structure.

There are three types of RPC endpoints:
1. Synchronous methods
2. Subscriptions to asynchronous notifications
3. Requests to unsubscribe from existing subscription

Endpoints are associated with namespaces. For example, UI endpoints are located
in "ui" namespace.

#### Conventions

* Response field names should apply the `lowerCamelCase` convention.

#### Call to synchronous method

*Request template*:
```json
{
	"jsonrpc": "2.0",
	"id": "uniqueStringOrNumber",
	"method": "namespaceName_methodName",
	"params": [ "paramValue1", "paramValue2", "..." ]
}
```

*Response template*:

```json
{
	"jsonrpc": "2.0",
	"id": "requestId",
	"result": "jsonObject",
	"error": { "code": "errorCode", "message": "errorMessage" }
}
```

#### Subscribe to asynchronous notifications

*Request template*:

```json
{
	"jsonrpc": "2.0",
	"id": "uniqueStringOrNumber",
	"method": "namespaceName_subscribe",
	"params": [ "subscriptionType", "paramValue1", "paramValue2", "..." ]
}
```

*Response template*:

```json
{
	"jsonrpc": "2.0",
	"id": "requestId",
	"result": "subscriptionId"
}
```

*Notification template*:

```json
{
	"jsonrpc": "2.0",
	"method": "namespaceName_subscription",
	"params": {
		"subscription": "subscriptionId",
		"result": "jsonObject"
	}
}
```

#### Unsubscribe from a given subscription

*Request template*:

```json
{
	"jsonrpc": "2.0",
	"id": "uniqueStringOrNumber",
	"method": "namespaceName_unsubscribe",
	"params": [ "subscriptionId" ]
}
```

*Response template*:

```json
{
	"jsonrpc": "2.0",
	"id": "requestId",
	"result": true
}
```
