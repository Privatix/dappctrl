from dappctrl_rpc import *

token = get_token(default_password)

accounts = get_accounts(token)
for account in accounts:
    print("\nProcessing account: {} ({})".format(account["name"], account["id"]))
    transfer_tokens(token, account["id"], account["ptcBalance"], "psc")
