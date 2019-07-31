import base64
import os

from dappctrl_rpc import *

private_key_file_name = "private_key.json"

set_password(default_password)

token = get_token(default_password)

account_id = create_account(token)
print("\tAccount id: {}".format(account_id))

account = get_object(token, "account", account_id)
print("\tEth address: 0x{}".format(account["ethAddr"]))

encoded_private_key = export_private_key(token, account["id"])
private_key = base64.b64decode(encoded_private_key)
print("\tPrivate key: {}".format(private_key))

with open(private_key_file_name, 'w') as f:
    f.write(private_key)

print("\tPrivate key file: {}/{}".format(
    os.path.dirname(os.path.realpath(__file__)),
    private_key_file_name))
