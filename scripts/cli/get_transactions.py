from dappctrl_rpc import *

token = get_token(default_password)

accounts = get_accounts(token)
for account in accounts:
    transactions = get_eth_transactions(token, "accountAggregated", account["id"], 0, 100)
    for transaction in transactions:
        print("-" * 80)
        print("\nTransaction: {}:\n\tStatus: {}\n\tIssued: {}\n\thttps://etherscan.io/tx/0x{}".format(
            transaction["method"],
            transaction["status"],
            transaction["issued"],
            transaction["hash"]))
