# Monitor

Monitor performs Log collecting periodically.

Several most recent blocks on the blockchain are considered `unreliable` (the relevant setting is `eth.min.confirmations`).

Let:
* A = last processed block number
* Z = most recent block number on the blockchain
* C = the min confirmations setting
* F = the fresh blocks setting
* L = the limit blocks to retrieve setting

Thus the range of interest it is:

```
    if F > 0 Ro = Ri ∩ [Z - C - F, +inf)
    else Ro = Ri
```

and:

```
    if L > 0 & (Z - C) > (A + 1) & (Z - C) - (A + 1) > L Ri = A + 1 + L
    else Ri = Z - C
```

These are the rules for filtering logs on the blockchain:

1. Events for Agent
  * Topics[0]: any
  * Topics[1]: one of accounts with `in_use = true`
1. Events for Client
  * Topics[0]: one of these hashes
    * LogChannelCreated
    * LogChannelToppedUp
    * LogChannelCloseRequested
    * LogCooperativeChannelClose
    * LogUnCooperativeChannelClose
    * Topics[2]: one of the accounts with `in_use = true`
1. Offering events
  * Topics[0]: one of these hashes
    * LogOfferingCreated
    * LogOfferingDeleted
    * LogOfferingPopedUp
  * Topics[1]: not one of the accounts with `in_use = true`
  * Topics[2]: one of the accounts with `in_use = true`
