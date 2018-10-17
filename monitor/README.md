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

For Agents:
1. all events with agent account address as topic[1]
  * Topics[0]: any
  * Topics[1]: one of accounts with `in_use = true`
2. all incoming transfers
  * Topics[0]: Transfer
  * Topics[2]: one of accounts with `in_use = true`

For Clients:
1.
  * Topics[0]: one of these hashes
    * Transfer
    * LogChannelToppedUp
    * LogChannelCloseRequested
  * Topics[2]: one of the accounts with `in_use = true`
2.
  * Topics[0]: one of these hashes
    * Tranfer
    * Approval
  * Topics[1]: one of the accounts with `in_use = true`
3.
  * Topics[0]: one of these hashes
    * LogChannelCreated
    * LogOfferingCreated
    * LogOfferingDeleted
    * LogOfferingPopedUp
    * LogCooperativeChannelClose
    * LogUnCooperativeChannelClose
 