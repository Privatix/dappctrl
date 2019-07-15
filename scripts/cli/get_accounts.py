import json

from dappctrl_rpc import *

token = get_token(default_password)

accounts = get_accounts(token)
for account in accounts:
    print("-" * 80)
    print(
        "\n{}:\n\tAddr: 0x{}\n\tETH: {}\n\tAccount: {} PRIX\n\tMarketplace: {} PRIX\n\n\tLast check: {}\n\tIn use: {}\n\tIs default: {}\n\tId: {}".format(
            account["name"],
            account["ethAddr"],
            eth(int(account["ethBalance"])),
            prix(int(account["ptcBalance"])),
            prix(int(account["pscBalance"])),
            account["lastBalanceCheck"],
            account["inUse"],
            account["isDefault"],
            account["id"],
        ))
