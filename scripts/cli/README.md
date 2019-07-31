# Tools for step-by-step offering publications

Prerequisites:

* python 3+

## Offering's publication steps

### 0. Install dependencies

```bash
./install_dependencies.sh
```

### 1. Create an account

Set up password for the account:

```bash
export DAPP_PASSWORD=some_password
```

This password will be used as:

* Application password
* Passphrase for a private key encryption

```bash
python create_account.py
```

### 2. Transfer to them eth and PRIX

By using exchange or own wallet.

### 3. Check that founds has been delivered

```bash
python update_balance.py
python get_accounts.py
```

### 4. Transfer all PRIX from the Account to the Marketplace 

```bash
python transfer_all_to_marketplace.py
```

Ensure, that PRIX has been transferred to the Marketplace 
(usually it takes 5-10 min):

```bash
python get_transactions.py
python get_accounts.py
```

### 5. Publish an offering

```bash
python publish_offering.py ./offering.json
```

Ensure, that the offering has been published (usually it takes 5-10 min):

```bash
python get_offerings.py
```

### 6. Transfer all earned PRIX from the Marketplace to the Account  

```bash
python transfer_all_to_account.py
```

Ensure, that PRIX has been transferred to the Account 
(usually it takes 5-10 min):

```bash
python get_transactions.py
python get_accounts.py
```

## Tools

### create_account.py

#### Usage

```bash
python create_account.py
```

#### Output

```
Get token
	Ok: <Response [200]>
	
Generate account
	Ok: <Response [200]>
	Account: eec83276-bc94-4dc4-b04f-cc5e5173a6fb
	
Get eth address
	Ok: <Response [200]>
	Eth address: 0x9486205adc7147ae551804c97c5bbb723ec7b826
	
Export private key
	Ok: <Response [200]>
	Private key: Private key: {"address":"9486205adc7147ae551804c97c5bbb723ec7b826","crypto":{"cipher":"aes-128-ctr","ciphertext":"92e1decf5c1ed689ef2cc2da3ec34f3bef0f7e7e49609265e5cfe1070afa86e7","cipherparams":{"iv":"4420d0d8232258bd4436cd57dc4dfcea"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"78b76e8040063c972eb2121e56d00df2eb8b538c9259dee7dbc9ee107d4970d0"},"mac":"cb2c3f9e53880ccb848dff1a6aa53bbaf5f79dbcb6c14205a5ac08b6bf1feb73"},"id":"01733e41-2e1c-4dc9-8a2e-50cf78485922","version":3}
	Private key file: /Users/user/tmp/private_key.json
```

### get_accounts.py

#### Usage

```bash
python get_accounts.py
```

#### Output

```
Get token
	Ok: <Response [200]>

Get account
	Ok: <Response [200]>
	Account: [
        {
                "ethAddr": "9486205adc7147ae551804c97c5bbb723ec7b826", 
                "name": "main", 
                "inUse": true, 
                "ptcBalance": 1000000000, 
                "lastBalanceCheck": "2019-06-19T17:11:57.227578+03:00", 
                "ethBalance": 50000000000000000, 
                "pscBalance": 0, 
                "id": "eec83276-bc94-4dc4-b04f-cc5e5173a6fb", 
                "isDefault": true
        }
]
```

### transfer_all_to_psc.py

#### Usage

```bash
python transfer_all_to_psc.py
```

#### Output

```
Get token
	Ok: <Response [200]>

Get accounts
	Ok: <Response [200]>

Processing account: main (eec83276-bc94-4dc4-b04f-cc5e5173a6fb)

Transfer tokens (amount: 1000000000, gas price: 6000000000, direction: psc)
	Ok: <Response [200]>
```

### get_transactions.py

#### Usage

```bash
python get_transactions.py
```

#### Output

```
Get token
	Ok: <Response [200]>

Get accounts
	Ok: <Response [200]>

Get eth transactions (type: accountAggregated, id: eec83276-bc94-4dc4-b04f-cc5e5173a6fb, offset: 0, limit: 100)
	Ok: <Response [200]>
--------------------------------------------------------------------------------

PSCAddBalanceERC20:
	sent 2019-06-20T12:10:30.705074+03:00
	https://etherscan.io/tx/0x91ad110fbb3ff0f2e32b7150d36ca6b1c9e8198b9da2561decb6c71933d3435c

PTCIncreaseApproval:
	sent 2019-06-20T12:10:01.500787+03:00
	https://etherscan.io/tx/0xc5bb8da80d7f68e1637c6d455e11a8c3cf6316d095dfd30fe09e1912d1c34a3e
```


### get_offerings.py

#### Usage

```bash
python get_offerings.py
```

#### Output

```
Get token
	Ok: <Response [200]>

Get products
	Ok: <Response [200]>

Get agent offerings (product_id: 89e338bf-f594-4c6d-89fc-6ccda002cf26, status: ['empty', 'registering', 'registered', 'popping_up', 'popped_up', 'removing', 'removed'], offset: 0, limit: 100)
	Ok: <Response [200]>

	VPN (ec4a47f7-3c56-4cd3-b3c4-41553f3cf6f1):
		status: registered
		hash: d902ddbcdefa0c924bee6a41825c19b0cd00e3153c68abbd930e43e7b00401d8
		supply: 30
		currentSupply: 30

	VPN (11400304-a20c-4348-aa88-999d2d309631):
		status: empty
		hash: 1efb5586c5a6506047e555cfe0126fb076a7a7b3ae8d65fff0c71772dc04a98c
		supply: 30
		currentSupply: 30
```


### publish_offering.py

#### Usage

```bash
python publish_offering.py offering.json
```

#### Output

```
Get token
	Ok: <Response [200]>

Get products
	Ok: <Response [200]>

Get accounts
	Ok: <Response [200]>

Used product: VPN

Used account: main

Offering: {
        "billingType": "postpaid", 
        "maxInactiveTimeSec": 1800, 
        "autoPopUp": true, 
        "description": "VPN", 
        "unitName": "MB", 
        "unitPrice": 1000, 
        "maxBillingUnitLag": 100, 
        "supply": 30, 
        "freeUnits": 0, 
        "agent": "eec83276-bc94-4dc4-b04f-cc5e5173a6fb", 
        "maxSuspendTime": 1800, 
        "product": "89e338bf-f594-4c6d-89fc-6ccda002cf26", 
        "billingInterval": 1, 
        "unitType": "units", 
        "serviceName": "VPN", 
        "template": "ab6964b8-5586-4bed-a546-795e944af586", 
        "minUnits": 10000, 
        "additionalParams": {
                "minDownloadMbits": 100, 
                "minUploadMbits": 80
        }, 
        "country": "RU", 
        "setupPrice": 0, 
        "maxUnit": 30000
}

Create offering
	Ok: <Response [200]>

Offering id: b706c092-dbcf-4847-b0f6-cf6c095d2cdd

Change offering status (offering_id: b706c092-dbcf-4847-b0f6-cf6c095d2cdd, action: publish)
	Ok: <Response [200]>
```

## get_errors.py

#### Usage

Get errors for the last 330 minutes:

```bash
python get_errors.py 330
```

#### Output

```
Get token
	Ok: <Response [200]>

Get logs (levels: ['error'], text: "", lower_bound: 2019-06-27T17:11:07.342929, upper_bound: 2019-06-27T22:11:07.342929, offset: 0, limit: 100)
	Ok: <Response [200]>
--------------------------------------------------------------------------------

error:
	job agentPreOfferingMsgBCPublish is failed

	{"job": "2c5e8877-4085-4dda-99f3-1992e2ef9939", "type": "agentPreOfferingMsgBCPublish", "method": "processWorker"}
--------------------------------------------------------------------------------

error:
	failed to register service offering: insufficient funds for gas * price + value

	{"GasLimit": 200000, "job": {"Status": "active", "NotBefore": "2019-06-27T17:40:52.172886Z", "Data": "eyJHYXNQcmljZSI6NjAwMDAwMDAwMH0=", "RelatedType": "offering", "CreatedBy": "user", "TryCount": 2, "RelatedID": "89763b49-5ef3-4467-8f9a-339544e0ed5e", "Type": "agentPreOfferingMsgBCPublish", "ID": "2c5e8877-4085-4dda-99f3-1992e2ef9939", "CreatedAt": "2019-06-27T17:38:39.164401Z"}, "type": "proc/worker.Worker", "method": "AgentPreOfferingMsgBCPublish", "GasPrice": 6000000000}
--------------------------------------------------------------------------------

```

## get_settings.py

#### Usage

Update settings:

```bash
python get_settings.py
```

#### Output

```
Get token
	Ok: <Response [200]>

Get settings
	Ok: <Response [200]>

eth.event.freshblocks: 11520
	Permissions: readWrite

updateDismissVersion: 
	Permissions: readWrite

eth.event.lastProcessedBlock: 8045538
	Permissions: readOnly

eth.max.deposit: 30000000000
	Permissions: readWrite

eth.event.blocklimit: 500
	Permissions: readWrite

error.sendremote: true
	Permissions: readWrite

```

## update_settings.py

#### Usage

Set `offering.autopopup` setting to `true`:

```bash
python update_settings.py '{"offering.autopopup": "true"}'
```

#### Output

```
Get token
	Ok: <Response [200]>

Update settings (settings: {u'offering.autopopup': u'false'})
	Ok: <Response [200]>
```

## remove_offering.py

#### Usage

Remove offering `89763b49-5ef3-4467-8f9a-339544e0ed5e`

```bash
python remove_offering.py '89763b49-5ef3-4467-8f9a-339544e0ed5e'
```

#### Output

```
Get token
	Ok: <Response [200]>

Change offering status (offering_id: 89763b49-5ef3-4467-8f9a-339544e0ed5e, action: deactivate)
	Ok: <Response [200]>
```

## popup_offering.py

#### Usage

Popup offering `89763b49-5ef3-4467-8f9a-339544e0ed5e`

```bash
python popup_offering.py '89763b49-5ef3-4467-8f9a-339544e0ed5e'
```

#### Output

```
Get token
	Ok: <Response [200]>

Change offering status (offering_id: 89763b49-5ef3-4467-8f9a-339544e0ed5e, action: popup)
	Ok: <Response [200]>
```
