import json
import sys

from dappctrl_rpc import *

offering_file_name = sys.argv[1]

token = get_token(default_password)

products = get_products(token)
product = products[0]

accounts = get_accounts(token)
account = accounts[0]

print("\nUsed product: {}".format(product["name"]))
print("\nUsed account: {}".format(account["name"]))

with open(offering_file_name) as f:
    offering = json.load(f)

offering["country"] = product["country"]
offering["product"] = product["id"]
offering["template"] = product["offerTplID"]
offering["agent"] = account["id"]
offering["serviceName"] = product["name"]
offering["description"] = product["name"]

offering["unitPrice"] = int(raw_prix(float(offering["unitPrice"])))
offering["setupPrice"] = int(raw_prix(float(offering["setupPrice"])))

# if maxUnit==0 it means the offering is unlimited
if offering["maxUnit"] == 0:
    del offering["maxUnit"]

print("\nOffering: {}".format(json.dumps(offering, indent=8)))

offering_id = create_offering(token, offering)
print("\nOffering id: {}".format(offering_id))

change_offering_status(token, offering_id, "publish")
