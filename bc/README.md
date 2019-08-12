# Monitor

Monitor performs Log collecting periodically.

Several most recent blocks on the blockchain are considered `unreliable` (the relevant setting is `eth.min.confirmations`).

Number of blocks to filter within one request is used to split total range in sub ranges (the relevant setting is `eth.event.blocklimit`).

On clients additional backward search is done to get offerings. The number of blocks to see back is used (the relevant setting is `eth.event.offeringsfreshblocks`).

## Forward search range of interest (Agent & Client)
Let:
* A = last processed block number
* Z = most recent block number on the blockchain
* C = the min confirmations setting
* L = the limit blocks to retrieve setting

Thus the range of interest it is:
```
if (A + 1) < (A + 1 + min(L, Z-C)) => [A+1, A + 1 + min(L, Z-C) ]
else range of interest is empty
```

## Backward search for offerings (Client only)
Let:
* A = last block number the last processed range started from
* Z = most recent block number on the blockchain
* C = the min confirmations setting
* L = the limit blocks to retrieve setting
* F = the offerings fresh block setting

Thus the range of interest for backward offerings earch is:
```
if (A - L - 1) < (A - 1) => [A - L - 1; A - 1]
else range of interest is empty
```

## Filtering rules
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
 