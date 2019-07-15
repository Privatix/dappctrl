import sys

from dappctrl_rpc import *

offering_id = sys.argv[1]

token = get_token(default_password)

change_offering_status(token, offering_id, "popup")
