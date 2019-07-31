from dappctrl_rpc import *

token = get_token(default_password)

accounts = get_accounts(token)
for account in accounts:
    update_balance(token, account["id"])
