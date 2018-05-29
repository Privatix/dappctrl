# Monitor

Monitor performs Log collecting periodically.

Several most recent blocks on the blockchain are considered `unreliable` (the relevant setting is `eth.min.confirmations`).

Let:
* A = last processed block number
* Z = most recent block number on the blockchain
* C = the min confirmations setting
* F = the fresh offerings setting

Thus the range of interest for agent and client logs Ri = [A + 1, Z - C],
and for offering it is:

```
    if F > 0 Ro = Ri ∩ [Z - C - F, +inf)
    else Ro = Ri
```

These are the rules for filtering logs on the blockchain:

1. Events for Agent
  * From: A + 1
  * To:   Z - C
  * Topics[0]: any
  * Topics[1]: one of accounts with `in_use = true`
1. Events for Client
  * From: A + 1
  * To:   Z - C
  * Topics[0]: one of these hashes
    * LogChannelCreated
    * LogChannelToppedUp
    * LogChannelCloseRequested
    * LogCooperativeChannelClose
    * LogUnCooperativeChannelClose
    * Topics[2]: one of the accounts with `in_use = true`
1. Offering events
  * From: max(A + 1, Z - C - F) if `F > 0`
  * From: A + 1 if `F == 0`
  * To:   Z - C
  * Topics[0]: one of these hashes
    * LogOfferingCreated
    * LogOfferingDeleted
    * LogOfferingPopedUp
  * Topics[1]: not one of the accounts with `in_use = true`
  * Topics[2]: one of the accounts with `in_use = true`
