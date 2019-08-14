from dappctrl_rpc import *

token = get_token(default_password)

suggested_gas_price = get_suggested_gas_price(token)
print("\nSuggested gas price: {}".format(eth_to_gwei(suggested_gas_price)))

accounts = get_accounts(token)
for account in accounts:
    print("\nProcessing account: {} ({})".format(account["name"], account["id"]))
    transfer_tokens(token, account["id"], account["pscBalance"], "ptc", suggested_gas_price)
