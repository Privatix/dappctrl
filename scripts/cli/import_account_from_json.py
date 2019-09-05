import json
import sys

from dappctrl_rpc import *

json_file_path = sys.argv[1]

with open(json_file_path) as f:
    json_content = json.load(f)

token = get_token(default_password)

account_params = {
    "isDefault": True,
    "name": "imported",
    "inUse": True
}

account_id = import_account_from_json(token, account_params, json_content, default_password)
print("\n\nAccount id: {}".format(account_id))
