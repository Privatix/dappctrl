import sys

from dappctrl_rpc import *

offering_id = sys.argv[1]

token = get_token(default_password)

suggested_gas_price = get_suggested_gas_price(token)
print("\nSuggested gas price: {}".format(eth_to_gwei(suggested_gas_price)))

change_offering_status(token, offering_id, "popup", suggested_gas_price)
