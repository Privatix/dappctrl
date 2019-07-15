from dappctrl_rpc import *

token = get_token(default_password)

products = get_products(token)
for product in products:
    offerings = get_agent_offerings(token, product["id"],
                                    ["empty", "registering", "registered",
                                     "popping_up", "popped_up", "removing",
                                     "removed"]
                                    , 0, 100)

    for offering in offerings:
        print("-" * 80)
        print("\n{}:\n\tHash: 0x{}\n\tStatus: {}\n\tSupply: {}\n\tCurrent supply: {}\n\tId: {}".format(
            offering["serviceName"],
            offering["hash"],
            offering["status"],
            offering["supply"],
            offering["currentSupply"],
            offering["id"],
        ))

        transactions = get_eth_transactions(token, "offering", offering["id"], 0, 100)
        for transaction in transactions:
            print("\n\tTransaction: {}:\n\t\tStatus: {}\n\t\tIssued: {}\n\t\thttps://etherscan.io/tx/0x{}".format(
                transaction["method"],
                transaction["status"],
                transaction["issued"],
                transaction["hash"]))
